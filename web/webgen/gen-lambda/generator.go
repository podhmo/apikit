package genlambda

import (
	"context"
	"fmt"
	"io"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/emitfile"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/translate"
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

	g.Emitter.FileEmitter.Config = &emitfile.Config{
		Verbose: g.Verbose,
		Log:     g.Log,
	}
	return g
}

type Generator struct {
	*Config

	Emitter *emitgo.Emitter
	Tracker *resolve.Tracker

	RootPkg *tinypkg.Package
}

func (g *Generator) Generate(
	ctx context.Context,
	pkg *tinypkg.Package,
	actionfunc interface{},
) error {
	t := translate.NewTranslator(g.Config.Config)

	// TODO: merge files

	here := pkg
	if g.Verbose {
		g.Log.Printf("\t+ generate handler package")
	}

	// event
	{
		c := t.TranslateToStruct(here, actionfunc, "Event")
		g.Emitter.Register(here, c.Name, c)
	}

	// handler
	{
		c := &code.CodeEmitter{Code: g.Config.NewCode(here, "Handle", func(w io.Writer, c *code.Code) error {
			fmt.Fprintf(w, "func Handle(ctx context.Context, event Event) (interface{}, error) {\n")
			defer fmt.Fprintf(w, "}\n")

			fmt.Fprintf(w, "\treturn nil, nil\n")
			return nil
		})}
		g.Emitter.Register(here, c.Name, c)
	}

	return nil
}
