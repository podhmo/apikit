//go:build apikit
// +build apikit

package main

import (
	"context"
	"fmt"
	"log"

	"m/13openapi/design"
	"m/13openapi/myplugins/gendoc"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/web"
	genchi "github.com/podhmo/apikit/web/webgen/gen-chi"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

// TODO: set error handler (500-handler)
// TODO: set 404-handler

func run() (err error) {
	emitter := emitgo.NewConfigFromRelativePath(design.ListArticle, "..").NewEmitter()
	defer emitter.EmitWith(&err)

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
	if err := g.Generate(
		context.Background(),
		r,
		design.HTTPStatusOf,
	); err != nil {
		return err
	}

	// generate openapi doc via custom plugin
	if err := g.IncludePlugin(g.RootPkg, gendoc.Options{
		OutputFile: "docs/openapi.json",
		Handlers:   g.Handlers,
	}); err != nil {
		return fmt.Errorf("on gendoc plugin: %w", err)
	}

	return nil
}
