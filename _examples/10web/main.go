package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"m/10web/design"

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

func run() (err error) {
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

	{
		resolver := translator.Config.Resolver
		tracker := translator.Tracker
		runtime := resolver.NewPackage("m/runtime", "")
		main := resolver.NewPackage("m/main", "")

		w := os.Stdout
		here := main

		getProviderModule, err := translator.GetProviderModule(here, "Provider")
		if err != nil {
			return err
		}
		runtimeModule, err := translator.RuntimeModule(runtime)
		if err != nil {
			return err
		}
		return web.Walk(r, func(node *web.WalkerNode) error {
			def := resolver.Def(node.Node.Value)
			tracker.Track(def)
			pathinfo, err := web.ExtractPathInfo(node.Node.VariableNames, def)
			if err != nil {
				return fmt.Errorf("extract path info: %w", err)
			}
			if err := webtranslate.WriteHandlerFunc(w, here, resolver, tracker, pathinfo, getProviderModule, runtimeModule, ""); err != nil {
				return fmt.Errorf("write: %w", err)
			}
			return nil
		})
	}
}
