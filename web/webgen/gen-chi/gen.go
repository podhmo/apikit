package genchi

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/web"
)

func New(emitter *emitgo.Emitter, translator *Translator) *Generator {
	rootpkg := emitter.RootPkg

	g := &Generator{
		RootPkg:    rootpkg,
		Emitter:    emitter,
		Translator: translator,
		Config:     translator.Config,
	}

	g.RuntimePkg = rootpkg.Relative("runtime", "")
	g.HandlerPkg = rootpkg.Relative("handler", "")
	g.ProviderPkg = g.HandlerPkg
	g.RouterPkg = g.HandlerPkg
	return g
}

type Generator struct {
	Emitter    *emitgo.Emitter
	Translator *Translator
	Config     *Config

	RootPkg     *tinypkg.Package
	ProviderPkg *tinypkg.Package
	RouterPkg   *tinypkg.Package
	HandlerPkg  *tinypkg.Package
	RuntimePkg  *tinypkg.Package
}

func (g *Generator) Generate(ctx context.Context, r *web.Router) error {
	c := g.Config
	c.RuntimePkg = g.RuntimePkg // TODO: remove
	c.ProviderPkg = g.ProviderPkg

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
			code := g.Translator.TranslateToHandler(here, node, "")
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
				providerModule, err := g.Translator.ProviderModule()
				if err != nil {
					return err
				}
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
		code := g.Translator.TranslateToInterface(here, name)
		g.Emitter.Register(here, code.Name, code)
	}
	return nil
}
