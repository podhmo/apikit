// +build apikit

package main

import (
	"context"
	"log"

	"m/11web/design"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/web"
	"github.com/podhmo/apikit/ext/scroll"
	genchi "github.com/podhmo/apikit/web/webgen/gen-chi"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

// TODO: parameters bindings
// TODO: set lift function
// TODO: set error handler (500-handler)
// TODO: set 404-handler

// TODO: gentle message ( extract path info: expected variables are [], but want variables are [articleId] (in def GetArticle): mismatch-number-of-variables)

func run() (err error) {
	emitter := emitgo.NewConfigFromRelativePath(design.ListArticle, "..").NewEmitter()
	emitter.FilenamePrefix = "gen_" // generated file name is "gen_<name>.go"
	defer emitter.EmitWith(&err)

	r := web.NewRouter()
	r.Group("/articles", func(r *web.Router) {
		// TODO: set tag

		r.Get("/", design.ListArticle)
		r.Get("/{articleId}", design.GetArticle)
		r.Post("/{articleId}/comments", design.PostArticleComment)
	})

	c := genchi.DefaultConfig()
	c.Override("db", design.NewDB)

	g := c.New(emitter)
	if err := g.Generate(
		context.Background(),
		r,
		design.HTTPStatusOf,
	); err != nil {
		return err
	}
	return g.IncludePlugin(g.RuntimePkg, scroll.Options{LatestIDTypeZeroValue: 0}) // latestIDType is int
}
