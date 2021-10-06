package genchi

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/podhmo/apikit/pkg/difftest"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
)

type Article struct{}
type DB struct{}

func Ping() (interface{}, error) {
	return map[string]interface{}{"message": "hello"}, nil
}
func PingWithContext(ctx context.Context) (interface{}, error) {
	return map[string]interface{}{"message": "hello"}, nil
}
func Greeting(message string) (interface{}, error) {
	return map[string]interface{}{"message": message}, nil
}

func ListArticle(db *DB) ([]*Article, error) {
	return nil, nil
}
func ListArticleWithContext(ctx context.Context, db *DB) ([]*Article, error) {
	return nil, nil
}

func TestWriteHandlerFunc(t *testing.T) {
	handlerName := "Handler"

	config := DefaultConfig()
	config.Header = ""
	resolver := config.Resolver

	main := resolver.NewPackage("main", "")
	config.RuntimePkg = resolver.NewPackage("m/runtime", "")
	config.ProviderPkg = main

	cases := []struct {
		msg      string
		here     *tinypkg.Package
		mount    func(r *web.Router)
		override func(t *resolve.Tracker)
		want     string
	}{
		{
			msg:   "no-deps",
			here:  main,
			mount: func(r *web.Router) { r.Get("/ping", Ping) },
			want: `package main

import (
	"github.com/podhmo/apikit/web/genchi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		result, err := genchi.Ping()
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},
		{
			msg:   "bind-path",
			here:  main,
			mount: func(r *web.Router) { r.Get("/greet/{message}", Greeting) },
			want: `package main

import (
	"github.com/podhmo/apikit/web/genchi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		message := runtime.PathParam(req, "message")
		result, err := genchi.Greeting(message)
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},
		// TODO: ng 404
		// TODO: path validation
		{
			msg:   "single-dep",
			here:  main,
			mount: func(r *web.Router) { r.Get("/articles", ListArticle) },
			want: `package main

import (
	"github.com/podhmo/apikit/web/genchi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err)
			return
		}
		var db *genchi.DB
		{
			db = provider.DB()
		}
		result, err := genchi.ListArticle(db)
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},
		{
			msg:   "single-dep-with-error",
			here:  main,
			mount: func(r *web.Router) { r.Get("/articles", ListArticle) },
			override: func(tracker *resolve.Tracker) {
				tracker.Override(reflect.TypeOf(&DB{}), "db", resolver.Def(func() (*DB, error) { return nil, nil }))
			},
			want: `package main

import (
	"github.com/podhmo/apikit/web/genchi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err)
			return
		}
		var db *genchi.DB
		{
			var err error
			db, err = provider.DB()
			if err != nil {
				runtime.HandleResult(w, req, nil, err); return
			}
		}
		result, err := genchi.ListArticle(db)
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},
		// with-context
		{
			msg:  "no-dep-with-context",
			here: main,
			mount: func(r *web.Router) {
				r.Get("/ping", PingWithContext)
			},
			want: `package main

import (
	"context"
	"github.com/podhmo/apikit/web/genchi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, _, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err)
			return
		}
		var ctx context.Context = req.Context()
		result, err := genchi.PingWithContext(ctx)
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},
		{
			msg:   "single-dep-with-context",
			here:  main,
			mount: func(r *web.Router) { r.Get("/articles", ListArticleWithContext) },
			want: `package main

import (
	"context"
	"github.com/podhmo/apikit/web/genchi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err)
			return
		}
		var ctx context.Context = req.Context()
		var db *genchi.DB
		{
			db = provider.DB()
		}
		result, err := genchi.ListArticleWithContext(ctx, db)
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},

		// TODO: unexpected action
		// TODO: path binding
		// TODO: handling error
	}
	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			translator := NewTranslator(config)

			r := web.NewRouter()
			c.mount(r)
			if c.override != nil {
				c.override(translator.Tracker)
			}

			if err := web.Walk(r, func(n *web.WalkerNode) error {
				code := translator.TranslateToHandler(c.here, n, handlerName)
				var buf strings.Builder
				if err := code.Emit(&buf); err != nil {
					t.Fatalf("unexpected error %+v", err)
				}
				if want, got := strings.TrimSpace(c.want), strings.TrimSpace(buf.String()); want != got {
					difftest.LogDiffGotStringAndWantString(t, got, want)
				}
				return nil
			}); err != nil {
				t.Fatalf("unexpected error %+v", err)
			}
		})
	}
}