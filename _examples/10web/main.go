package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"m/10routing/design"

	"github.com/podhmo/apikit/web"
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

func run() (err error) {
	r := web.NewRouter()
	r.Group("/articles", func(r *web.Router) {
		// TODO: set tag

		r.Get("/", design.ListArticle)
		r.Get("/{articleId}", design.GetArticle)
	})

	enc := json.NewEncoder(os.Stdout)
	web.Walk(r, func(n *web.WalkerNode) error {
		return enc.Encode(map[string]interface{}{
			"path": strings.Join(n.Path(), ""),
			"vars": n.Node.VariableNames,
		})
	})
	return nil
}
