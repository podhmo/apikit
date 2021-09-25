package webtranslate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
)

type Article struct{}
type DB struct{}

func listArticle(db *DB) ([]*Article, error) {
	return nil, nil
}

func TestWriteHandlerFUnc(t *testing.T) {
	r := web.NewRouter()
	r.Get("/articles/", listArticle)

	var node *web.WalkerNode
	web.Walk(r, func(n *web.WalkerNode) error {
		node = n
		return nil
	})

	resolver := resolve.NewResolver()
	def := resolver.Def(node.Node.Value)
	pathinfo, err := web.ExtractPathInfo(node.Node.VariableNames, def)
	if err != nil {
		t.Fatalf("unexpected error, extract info, %+v", err)
	}

	main := resolver.NewPackage("main", "")
	var buf strings.Builder
	WriteHandlerFunc(&buf, main, resolver, pathinfo)
	fmt.Println(buf.String())
}
