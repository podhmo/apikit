package genchi

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
)

type Config struct {
	*code.Config

	Tracker *resolve.Tracker

	ProviderName string
	Verbose      bool
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
	g.RuntimePkg = rootpkg.Relative("runtime", "")
	g.HandlerPkg = rootpkg.Relative("handler", "")
	g.ProviderPkg = g.HandlerPkg
	g.RouterPkg = g.HandlerPkg
	return g
}

func (g *Generator) Generate(
	ctx context.Context,
	r *web.Router,
	getHTTPStatusFromError func(error) int,
) error {
	if g.Verbose {
		log.Printf("* runtime package -> %s", g.RuntimePkg.Path)
		log.Printf("* handler package -> %s", g.HandlerPkg.Path)
		log.Printf("* provider package -> %s", g.ProviderPkg.Path)
		log.Printf("* router package -> %s", g.RouterPkg.Path)
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
		here := g.RouterPkg
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
		code := translator.TranslateToInterface(here, name)
		g.Emitter.Register(here, code.Name, code)
	}

	// runtime (copy)
	{
		here := g.RuntimePkg
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
		g.Emitter.Register(here, "HandleResult", &code.CodeEmitter{Code: g.Config.NewCode(here, "runtime", func(w io.Writer, c *code.Code) error {
			pkg := resolver.NewPackage(emitgo.PackagePath(getHTTPStatusFromError), "")
			c.Import(pkg)

			fmt.Fprintln(w, "func init(){")
			defer fmt.Fprintln(w, "}")

			fmt.Fprintf(w, "\tHandleResult = %s(%s)\n",
				createHandleResultFunc,
				here.Import(pkg).Lookup(pkg.NewSymbol(resolver.Shape(getHTTPStatusFromError).GetName())),
			)
			fmt.Fprintln(w, "")
			return nil
		})})
	}
	return nil
}
