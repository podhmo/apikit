package webtranslate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
)

type Article struct{}
type DB struct{}

func ListArticle(db *DB) ([]*Article, error) {
	return nil, nil
}

func TestWriteHandlerFUnc(t *testing.T) {
	r := web.NewRouter()
	r.Get("/articles/", ListArticle)

	var node *web.WalkerNode
	web.Walk(r, func(n *web.WalkerNode) error {
		node = n
		return nil
	})

	config := DefaultConfig()
	resolver := config.Resolver

	def := resolver.Def(node.Node.Value)
	pathinfo, err := web.ExtractPathInfo(node.Node.VariableNames, def)
	if err != nil {
		t.Fatalf("unexpected error, extract info, %+v", err)
	}

	main := resolver.NewPackage("main", "")
	runtime := resolver.NewPackage("m/runtime", "")

	var buf strings.Builder
	providerFunc := &tinypkg.Var{
		Name: "getProvider",
		Node: main.NewFunc(
			"getProvider",
			[]*tinypkg.Var{{Node: &tinypkg.Pointer{Lv: 1, V: resolve.NewResolver().NewPackage("net/http", "").NewSymbol("Request")}}},
			[]*tinypkg.Var{{Node: main.NewSymbol("Provider")}},
		)}
	if err := WriteHandlerFunc(&buf, main, "", resolver, pathinfo, runtime, providerFunc); err != nil {
		t.Errorf("unexpected error %+v", err)
	}
	fmt.Println(buf.String())
}
