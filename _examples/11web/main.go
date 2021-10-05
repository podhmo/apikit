package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"m/11web/design"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/web"
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
	rootpkg := emitter.RootPkg

	r := web.NewRouter()
	r.Group("/articles", func(r *web.Router) {
		// TODO: set tag

		r.Get("/", design.ListArticle)
		r.Get("/{articleId}", design.GetArticle)
	})

	enc := json.NewEncoder(os.Stdout)

	fmt.Println("----------------------------------------")
	web.Walk(r, func(n *web.WalkerNode) error {
		return enc.Encode(map[string]interface{}{
			"path": strings.Join(n.Path(), ""),
			"vars": n.Node.VariableNames,
		})
	})
	fmt.Println("----------------------------------------")

	translator := webtranslate.NewTranslator(webtranslate.DefaultConfig())
	resolver := translator.Resolver

	translator.Config.RuntimePkg = resolver.NewPackage("github.com/podhmo/apikit/web/webruntime", "") // xxx
	pkg := rootpkg.Relative("handler", "")
	// translator.Config.ProviderPkg = pkg

	translator.Override("db", design.NewDB)

	{
		here := pkg
		if err := web.Walk(r, func(node *web.WalkerNode) error {
			code := translator.TranslateToHandler(here, node, "")
			emitter.Register(here, code.Name, code)
			return nil
		}); err != nil {
			return err
		}
	}

	{
		here := pkg
		code := translator.TranslateToInterface(here)
		emitter.Register(here, code.Name, code)
	}
	return nil
}
