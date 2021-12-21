package main

import (
	"encoding/json"
	"os"
	"strings"

	"m/10routing/design"

	"github.com/podhmo/apikit/web"
)

func mount(r *web.Router) {
	r.Group("/articles", func(r *web.Router) {
		// TODO: set tag

		r.Get("/", design.ListArticle)
		r.Get("/{articleId}", design.GetArticle)
	})
}

func main() {
	r := web.NewRouter()
	mount(r)

	dumpJSON := json.NewEncoder(os.Stdout).Encode
	err := web.Walk(r, func(n *web.WalkerNode) error {
		return dumpJSON(map[string]interface{}{
			"path": strings.Join(n.Path(), ""),
			"vars": n.Node.VariableNames,
		})
	})
	if err != nil {
		panic(err)
	}
}
