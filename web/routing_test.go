package web_test

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/podhmo/apikit/pkg/difftest"
	"github.com/podhmo/apikit/web"
)

func TestRouting(t *testing.T) {
	cases := []struct {
		msg    string
		router *web.Router
		want   []string
	}{
		{
			msg: "one",
			router: func() *web.Router {
				r := web.NewRouter()
				r.Get("/articles/{articleId}", "get article")
				return r
			}(),
			want: []string{
				"GET /articles/{articleId}",
			},
		},
		{
			msg: "many",
			router: func() *web.Router {
				r := web.NewRouter()
				r.Get("/articles/{articleId}", "get article")
				r.Group("/articles/{articleId}", func(r *web.Router) {
					r.Get("/comments", "list comments")
					r.Get("/comments/{commentId}/", "get comment")
				})
				return r
			}(),
			want: []string{
				"GET /articles/{articleId}",
				"GET /articles/{articleId}/comments",
				"GET /articles/{articleId}/comments/{commentId}",
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			var paths []string
			if err := web.Walk(c.router, func(n *web.WalkerNode) error {
				paths = append(paths, strings.Join(n.Path(), ""))
				return nil
			}); err != nil {
				t.Fatalf("unexpected error %+v", err)
			}

			if want, got := c.want, paths; !reflect.DeepEqual(want, got) {
				sort.Strings(want)
				sort.Strings(got)
				difftest.LogDiffGotStringAndWantString(t, strings.Join(got, "\n"), strings.Join(want, "\n"))
			}
		})
	}
}
