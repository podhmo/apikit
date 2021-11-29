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
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/plugins"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/translate"
	"github.com/podhmo/apikit/web"
)

type Config struct {
	*code.Config
	Tracker      *resolve.Tracker
	ProviderName string
}

func DefaultConfig() *Config {
	c := code.DefaultConfig()
	c.Config.IgnoreMap["net/http.Request"] = true        // xxx:
	c.Config.IgnoreMap["net/http.ResponseWriter"] = true // xxx:
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
func (c *Config) New(emitter *emitgo.Emitter) *Generator {
	rootpkg := emitter.RootPkg
	g := &Generator{
		Emitter: emitter,
		Tracker: c.Tracker,
		Config:  c,
		RootPkg: rootpkg,
	}

	g.Emitter.FileEmitter.Config.Verbose = g.Verbose
	g.Emitter.FileEmitter.Config.Log = g.Log

	return g
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

// type GeneratorOption func(*Generator) error
// func (g *Generator) WithPlugin()

func (g *Generator) IncludePlugin(here *tinypkg.Package, plugin plugins.Plugin) error {
	// TODO: fix panic using before Generate()
	pc := &plugins.PluginContext{Config: g.Config.Config, Emitter: g.Emitter, Resolver: g.Resolver}
	return pc.IncludePlugin(here, plugin)
}

func (g *Generator) Generate(
	ctx context.Context,
	r *web.Router,
	getHTTPStatusFromError func(error) int,
) error {
	if g.HandlerPkg == nil {
		g.HandlerPkg = g.RootPkg.Relative("handler", "")
	}
	if g.RuntimePkg == nil {
		g.RuntimePkg = g.HandlerPkg.Relative("runtime", "")
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

	analyzer, err := newAnalyzer(g)
	if err != nil {
		return fmt.Errorf("new analyzer: %w", err)
	}

	resolver := g.Tracker.Resolver

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
			name := web.GetMetaData(node.Node).Name
			analyzed, err := analyzer.Analyze(here, node)
			// TODO: add hook
			if err != nil {
				return fmt.Errorf("analyze failure: %w", err)
			}
			code := ToHandlerCode(here, g.Config.Config, analyzed, name)
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
				providerModule := analyzer.ProviderModule
				c.AddDependency(providerModule)
				getProviderFunc := providerModule.Funcs.GetProvider

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

		translator := &translate.Translator{
			Tracker:  g.Tracker,
			Resolver: resolver,
			Config:   g.Config.Config,
		}
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

		// runtime.go
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
	}

	// runtime-handle (copy)
	{
		here := g.RuntimePkg

		pkg := resolver.NewPackage(emitgo.PackagePath(getHTTPStatusFromError), "")
		getStatusFunc := here.Import(pkg).Lookup(pkg.NewSymbol(resolver.Shape(getHTTPStatusFromError).GetName()))
		if g.Verbose {
			g.Log.Printf("\t+ generate runtime-handle [getStatus=%s]", getStatusFunc)
		}

		// handle.go
		c := &code.CodeEmitter{Code: g.Config.NewCode(here, "handle", func(w io.Writer, c *code.Code) error {
			fpath := filepath.Join(emitgo.DefinedDir(DefaultConfig), "webruntime/handle.go")
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

			runtimeModule, err := analyzer.runtimeModule.Imported(c.Here) // xxx
			if err != nil {
				return err
			}
			createHandleResultFunc := runtimeModule.Symbols.CreateHandleResult

			c.Import(pkg)
			fmt.Fprintln(w, "func init(){")
			fmt.Fprintf(w, "\tHandleResult = %s(%s)\n", createHandleResultFunc, getStatusFunc)
			fmt.Fprintln(w, "}")
			return nil
		})}
		g.Emitter.Register(here, c.Name, c)
	}
	return nil
}
