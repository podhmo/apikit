//go:build apikit
// +build apikit

package main

import (
	"context"

	"m/13openapi/design"
	"m/13openapi/design/enum"
	"m/13openapi/myplugins/gendoc"
	"m/13openapi/seed"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/web"
	genchi "github.com/podhmo/apikit/web/webgen/gen-chi"
)

func main() {
	ctx := context.Background()
	emitgo.NewConfigFromRelativePath(design.ListArticle, "..").MustEmitWith(func(emitter *emitgo.Emitter) error {
		c := genchi.DefaultConfig()
		override(c)

		r := web.NewRouter()
		mount(r)

		g := c.New(emitter)
		if err := g.Generate(ctx, r, design.HTTPStatusOf); err != nil {
			return err
		}

		// generate openapi doc via custom plugin
		type defaultError struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		return g.ActivatePlugins(ctx, g.RootPkg,
			gendoc.Options{
				OutputFile:   "docs/openapi.json",
				Handlers:     g.Handlers,
				DefaultError: defaultError{},
				Prepare: func(m *gendoc.Manager) {
					m.DefineEnumWithEnumSet(enum.SortOrderDesc, seed.Enums.SortOrder)
				},
			},
		)
	})
}

func mount(r *web.Router) {
	r.Group("/articles", func(r *web.Router) {
		// TODO: set tag
		r.Get("/", design.ListArticle)
		r.Get("/{articleId}", design.GetArticle)
		r.Post("/{articleId}/comments", design.PostArticleComment)
	})
}

func override(c *genchi.Config) {
	c.Override("db", design.NewDB)
}
