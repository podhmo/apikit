package genchi

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/emitfile"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
)

type Config struct {
	*code.Config
	Tracker      *resolve.Tracker
	ProviderName string
}

func DefaultConfig() *Config {
	c := code.DefaultConfig()
	return &Config{
		Config:       c,
		Tracker:      resolve.NewTracker(c.Resolver),
		ProviderName: "Provider",
	}
}

func (c *Config) Override(name string, fn interface{}) (*resolve.Def, error) {
	return c.Tracker.Override(name, fn)
}
func (c *Config) NewPackage(path, name string) *tinypkg.Package {
	return c.Resolver.NewPackage(path, name)
}

type Generator struct {
	*Config

	Emitter *emitgo.Emitter
	Tracker *resolve.Tracker

	RootPkg     *tinypkg.Package
	ProviderPkg *tinypkg.Package
	RouterPkg   *tinypkg.Package
	HandlerPkg  *tinypkg.Package
	RuntimePkg  *tinypkg.Package
}

func (c *Config) New(emitter *emitgo.Emitter) *Generator {
	rootpkg := emitter.RootPkg
	g := &Generator{
		Emitter: emitter,
		Tracker: c.Tracker,
		Config:  c,
		RootPkg: rootpkg,
	}

	g.Emitter.FileEmitter.Config = &emitfile.Config{
		Verbose: g.Verbose,
		Log:     g.Log,
	}
	return g
}

func (g *Generator) Generate(
	ctx context.Context,
	r *web.Router,
	getHTTPStatusFromError func(error) int,
) error {
	if g.RuntimePkg == nil {
		g.RuntimePkg = g.RootPkg.Relative("runtime", "")
	}
	if g.HandlerPkg == nil {
		g.HandlerPkg = g.RootPkg.Relative("handler", "")
	}
	if g.ProviderPkg == nil {
		g.ProviderPkg = g.HandlerPkg
	}
	if g.RouterPkg == nil {
		g.RouterPkg = g.HandlerPkg
	}

	g.Log.Printf("detect target packages ...")
	if g.Verbose {
		g.Log.Printf("\t* runtime package -> %s", g.RuntimePkg.Path)
		g.Log.Printf("\t* handler package -> %s", g.HandlerPkg.Path)
		g.Log.Printf("\t* provider package -> %s", g.ProviderPkg.Path)
		g.Log.Printf("\t* router package -> %s", g.RouterPkg.Path)
	}

	resolver := g.Tracker.Resolver
	providerModule, err := ProviderModule(g.ProviderPkg, resolver, g.Config.ProviderName)
	if err != nil {
		return err
	}
	runtimeModule, err := RuntimeModule(g.RuntimePkg, resolver)
	if err != nil {
		return err
	}
	createHandleResultFunc, err := runtimeModule.Symbol(g.RuntimePkg, "CreateHandleResultFunction")
	if err != nil {
		return err
	}

	translator := &Translator{
		Resolver:       resolver,
		Tracker:        g.Tracker,
		Config:         g.Config.Config,
		ProviderModule: providerModule,
		RuntimeModule:  runtimeModule,
	}

	type handler struct {
		name   string
		method string
		path   string
		fn     tinypkg.Node
	}
	var handlers []handler

	// handler
	{
		g.Log.Printf("generate handler package ...")
		here := g.RouterPkg

		if err := web.Walk(r, func(node *web.WalkerNode) error {
			code := translator.TranslateToHandler(here, node, "")
			g.Emitter.Register(here, code.Name, code)

			methodAndPath := strings.SplitN(strings.Join(node.Path(), ""), " ", 2)
			h := handler{
				name:   code.Name,
				method: methodAndPath[0][:1] + strings.ToLower(methodAndPath[0][1:]), // GET -> Get
				path:   methodAndPath[1],
				fn:     here.NewSymbol(code.Name),
			}
			handlers = append(handlers, h)
			return nil
		}); err != nil {
			return fmt.Errorf("on generate handler: %w", err)
		}
	}

	// routing
	// TODO: get provider func
	{
		g.Log.Printf("generate router package ...")
		here := g.RouterPkg
		if g.Verbose {
			g.Log.Printf("\t+ generate %s.Mount()", here.Path)
		}

		g.Emitter.Register(here, "mount.go", &code.CodeEmitter{Code: g.Config.NewCode(
			here, "Mount",
			func(w io.Writer, c *code.Code) error {
				c.AddDependency(providerModule)
				getProviderFunc, err := providerModule.Type("getProvider")
				if err != nil {
					return fmt.Errorf("in provider module, %w", err)
				}

				chi := g.Config.Resolver.NewPackage("github.com/go-chi/chi/v5", "chi")
				f := here.NewFunc("Mount", []*tinypkg.Var{
					{Name: "r", Node: chi.NewSymbol("Router")},
					{Name: getProviderFunc.Name, Node: getProviderFunc},
				}, nil)
				c.AddDependency(f)

				return tinypkg.WriteFunc(w, here, "", f, func() error {
					for _, h := range handlers {
						// TODO: grouping
						fmt.Fprintf(w, "\tr.%s(%q, %s(%s))\n", h.method, h.path, h.fn, getProviderFunc.Name)
					}
					return nil
				})
			},
		)})
	}

	// provider
	{
		here := g.ProviderPkg
		name := g.Config.ProviderName // xxx
		code := translator.ExtractProviderInterface(here, name)
		g.Emitter.Register(here, code.Name, code)
	}

	// runtime (copy)
	{
		g.Log.Printf("generate runtime package ...")
		here := g.RuntimePkg
		if g.Verbose {
			g.Log.Printf("\t+ generate runtime (almost copy)")
		}

		c := &code.CodeEmitter{Code: g.Config.NewCode(here, "runtime", func(w io.Writer, c *code.Code) error {
			fpath := filepath.Join(emitgo.DefinedDir(DefaultConfig), "webruntime/runtime.go")
			f, err := os.Open(fpath)
			if err != nil {
				return err
			}

			defer f.Close()
			r := bufio.NewReader(f)
			for {
				line, _, err := r.ReadLine()
				if err != nil {
					return err
				}
				if strings.HasPrefix(string(line), "package ") {
					break
				}
			}
			if _, err := io.Copy(w, r); err != nil {
				return err
			}
			return nil
		})}
		g.Emitter.Register(here, c.Name, c)

		// generate HandleResult = CreateGenerateHandleResult
		pkg := resolver.NewPackage(emitgo.PackagePath(getHTTPStatusFromError), "")
		getStatusFunc := here.Import(pkg).Lookup(pkg.NewSymbol(resolver.Shape(getHTTPStatusFromError).GetName()))
		if g.Verbose {
			g.Log.Printf("\t+ generate HandleResult() with %s", getStatusFunc)
		}
		g.Emitter.Register(here, "HandleResult", &code.CodeEmitter{Code: g.Config.NewCode(here, "runtime", func(w io.Writer, c *code.Code) error {
			c.Import(pkg)
			fmt.Fprintln(w, "func init(){")
			fmt.Fprintf(w, "\tHandleResult = %s(%s)\n", createHandleResultFunc, getStatusFunc)
			fmt.Fprintln(w, "}")
			return nil
		})})
	}
	return nil
}