package gengraphqlgo

import (
	"context"
	"io"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/graph"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/plugins"
	"github.com/podhmo/apikit/resolve"
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
	ResolverPkg *tinypkg.Package
}

// type GeneratorOption func(*Generator) error
// func (g *Generator) WithPlugin()

func (g *Generator) ActivatePlugins(ctx context.Context, here *tinypkg.Package, targets ...plugins.Plugin) error {
	// TODO: fix panic using before Generate()
	c := &plugins.PluginConfig{Config: g.Config.Config}
	c.Emitter = g.Emitter
	c.Resolver = g.Resolver
	return c.ActivatePlugins(ctx, here, targets...)
}

func (g *Generator) Generate(
	ctx context.Context,
	r *graph.Router,
) error {
	if g.ResolverPkg == nil {
		g.ResolverPkg = g.RootPkg.Relative("resolver", "")
	}
	if g.ProviderPkg == nil {
		g.ProviderPkg = g.ResolverPkg
	}

	g.Log.Printf("detect target packages ...")
	if g.Verbose {
		g.Log.Printf("\t* resolver package -> %s", g.ResolverPkg.Path)
		g.Log.Printf("\t* provider package -> %s", g.ProviderPkg.Path)
	}

	c := g.NewCode(g.ResolverPkg, "mount", func(w io.Writer, c *code.Code) error {
		return nil
	})
	g.Emitter.Register(c.Here, c.Name, &code.CodeEmitter{Code: c})

	return nil
}
