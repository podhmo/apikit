package main

import (
	"encoding/json"
	"log"
	"m/10web/design"
	"os"
	"strings"

	"github.com/podhmo/apikit/web"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() (err error) {
	r := web.NewRouter()
	r.Group("/articles", func(r *web.Router) {
		// TODO: set tag

		r.Get("/", design.ListArticle)
		r.Get("/{articleId}", design.GetArticle)
	})

	enc := json.NewEncoder(os.Stdout)
	return web.Walk(r, func(n *web.WalkerNode) error {
		return enc.Encode(map[string]interface{}{
			"path": strings.Join(n.Path(), ""),
			"vars": n.Node.VariableNames,
		})
	})
}
