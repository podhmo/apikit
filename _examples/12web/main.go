package main

import (
	"context"
	"log"

	"m/12web/design"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/web"
	genchi "github.com/podhmo/apikit/web/webgen/gen-chi"
	"github.com/podhmo/apikit/web/webtranslate"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

// TODO: code generation
// TODO: parameters bindings
// TODO: set lift function
// TODO: set error handler (500-handler)
// TODO: set 404-handler

// TODO: normalize argname and path parameter. (e.g. articleID and articleId)
// TODO: gentle message ( extract path info: expected variables are [], but want variables are [articleId] (in def GetArticle): mismatch-number-of-variables)
// - which is wrong? (routing?, func-def?)

// TODO: generate routing

func run() (err error) {
	emitter := emitgo.NewFromRelativePath(design.ListArticle, "..")
	defer emitter.EmitWith(&err)

	r := web.NewRouter()
	r.Group("/articles", func(r *web.Router) {
		// TODO: set tag

		r.Get("/", design.ListArticle)
		r.Get("/{articleId}", design.GetArticle)
	})

	translator := webtranslate.NewTranslator(webtranslate.DefaultConfig())
	translator.Override("db", design.NewDB)

	g := genchi.New(emitter, translator)
	g.RuntimePkg = translator.Resolver.NewPackage("github.com/podhmo/apikit/web/webruntime", "")
	return g.Generate(context.Background(), r)
}
