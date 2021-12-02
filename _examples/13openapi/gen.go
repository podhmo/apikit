//go:build apikit
// +build apikit

package main

import (
	"context"
	"fmt"
	"log"

	"m/13openapi/design"
	"m/13openapi/design/enum"
	"m/13openapi/myplugins/gendoc"
	"m/13openapi/seed"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/web"
	genchi "github.com/podhmo/apikit/web/webgen/gen-chi"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

// TODO: schema description
// TODO: (schema fields description)
// TODO: examples
// TODO: set error handler (500-handler)
// TODO: set 404-handler

func run() (err error) {
	ctx := context.Background()
	return emitgo.NewConfigFromRelativePath(design.ListArticle, "..").EmitWith(func(emitter *emitgo.Emitter) error {
		c := genchi.DefaultConfig()
		c.Override("db", design.NewDB)

		r := web.NewRouter()
		r.Group("/articles", func(r *web.Router) {
			// TODO: set tag
			r.Get("/", design.ListArticle)
			r.Get("/{articleId}", design.GetArticle)
			r.Post("/{articleId}/comments", design.PostArticleComment)
		})

		g := c.New(emitter)
		if err := g.Generate(ctx, r, design.HTTPStatusOf); err != nil {
			return err
		}

		// generate openapi doc via custom plugin
		type defaultError struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := g.IncludePlugin(g.RootPkg, gendoc.Options{
			OutputFile:   "docs/openapi.json",
			Handlers:     g.Handlers,
			DefaultError: defaultError{},
			Prepare: func(m *gendoc.Manager) {
				m.DefineEnumWithEnumSet(enum.SortOrderDesc, seed.Enums.SortOrder)
			},
		}); err != nil {
			return fmt.Errorf("on gendoc plugin: %w", err)
		}
		return nil
	})
}
