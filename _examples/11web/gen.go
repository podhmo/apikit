//go:build apikit
// +build apikit

package main

import (
	"context"

	"m/11web/design"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/plugins/scroll"
	"github.com/podhmo/apikit/web"
	genchi "github.com/podhmo/apikit/web/webgen/gen-chi"
)

func main() {
	ctx := context.Background()

	emitgo.NewConfigFromRelativePath(design.ListArticle, "..").MustEmitWith(func(emitter *emitgo.Emitter) error {
		emitter.FilenamePrefix = "gen_" // generated file name is "gen_<name>.go"

		r := web.NewRouter()
		mount(r)

		c := genchi.DefaultConfig()
		override(c)

		g := c.New(emitter)
		if err := g.Generate(ctx, r, design.HTTPStatusOf); err != nil {
			return err
		}
		return g.ActivatePlugins(ctx, g.RuntimePkg,
			scroll.Options{LatestIDTypeZeroValue: 0}, // latestIDType is int
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
