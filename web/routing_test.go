package web_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/podhmo/apikit/pkg/difftest"
	"github.com/podhmo/apikit/web"
)

func TestRouting(t *testing.T) {
	cases := []struct {
		msg           string
		router        *web.Router
		want          []string
		wantVariables []string
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
			wantVariables: []string{
				"articleId",
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
			wantVariables: []string{
				"articleId",
				"articleId",
				"articleId, commentId",
			},
		},
		{
			msg: "regex",
			router: func() *web.Router {
				r := web.NewRouter()
				r.Get("/{ articleSlug:[a-z-]+}", "getArticleBySlug")
				return r
			}(),
			want: []string{
				"GET /{articleSlug:[a-z-]+}",
			},
			wantVariables: []string{
				"articleSlug:[a-z-]+",
			},
		},
		{
			msg: "regression-1",
			router: func() *web.Router {
				r := web.NewRouter()
				r.Group("/articles", func(r *web.Router) {
					r.Get("/{articleId}", "GetArticle")
					r.Post("/{articleId}/comments", "PostArticleComment")
				})
				return r
			}(),
			want: []string{
				"POST /articles/{articleId}/comments",
				"GET /articles/{articleId}",
			},
			wantVariables: []string{
				"articleId",
				"articleId",
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			var paths []string
			var variables []string
			if err := web.Walk(c.router, func(n *web.WalkerNode) error {
				paths = append(paths, strings.Join(n.Path(), ""))
				variables = append(variables, strings.Join(n.Node.VariableNames, ", "))
				return nil
			}); err != nil {
				t.Fatalf("unexpected error %+v", err)
			}

			if want, got := c.want, paths; !reflect.DeepEqual(want, got) {
				difftest.LogDiffGotStringAndWantString(t, strings.Join(got, "\n"), strings.Join(want, "\n"))
			}

			if want, got := c.wantVariables, variables; !reflect.DeepEqual(want, got) {
				difftest.LogDiffGotStringAndWantString(t, strings.Join(got, "\n"), strings.Join(want, "\n"))
			}
		})
	}
}
