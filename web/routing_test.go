package web_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/podhmo/apikit/web"
)

func TestRouting(t *testing.T) {
	r := web.NewRouter()
	r.Method("GET", "/articles/{articleId}", "get article")
	r.Group("/articles/{articleId}", func(r *web.Router) {
		r.Method("GET", "/comments", "list comments")
		r.Method("GET", "/comments/{commentId}/", "get comment")
	})

	var paths []string
	if err := web.Walk(r, func(n *web.WalkerNode) error {
		paths = append(paths, strings.Join(n.Path(), ""))
		return nil
	}); err != nil {
		t.Fatalf("unexpected error %+v", err)
	}

	want := []string{
		"GET /articles/{articleId}",
		"GET /articles/{articleId}/{articleId}",
		"GET /articles/{articleId}/{articleId}/{commentId}",
	}

	if got := paths; !reflect.DeepEqual(want, got) {
		t.Errorf("want paths:\n\t%s\nbut got:\n\t%s", want, got)
	}
}
