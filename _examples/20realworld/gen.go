// +build apikit

package main

import (
	"context"
	"log"

	"m/action"
	"m/design"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/web"
	genchi "github.com/podhmo/apikit/web/webgen/gen-chi"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() (err error) {
	emitter := emitgo.NewFromRelativePath(action.CreateArticle, "..")
	defer emitter.EmitWith(&err)

	r := web.NewRouter()
	r.Post("/articles", action.CreateArticle)

	c := genchi.DefaultConfig()
	c.Verbose = true

	g := c.New(emitter)
	g.RuntimePkg = g.RootPkg.Relative("web/runtime", "")
	g.HandlerPkg = g.RootPkg.Relative("web/handler", "")

	return g.Generate(
		context.Background(),
		r,
		design.HTTPStatusOf,
	)
}
