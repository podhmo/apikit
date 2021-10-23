package genchi

import (
	"context"
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

type Message string

func GreetingWithNewType(message Message) (interface{}, error) {
	return map[string]interface{}{"message": message}, nil
}

func GreetingWithQueryString(message Message, verbose *bool) (interface{}, error) {
	return map[string]interface{}{"message": message}, nil
}

type Data struct {
	Message string `json:"message"`
}

func PostMessage(data Data) (interface{}, error) { return nil, nil }
func ListArticle(db *DB) ([]*Article, error) {
	return nil, nil
}
func ListArticleWithContext(ctx context.Context, db *DB) ([]*Article, error) {
	return nil, nil
}

func LoginRequired(db *DB) error {
	return nil
}

func TestWriteHandlerFunc(t *testing.T) {
	handlerName := "Handler"

	config := DefaultConfig()
	config.Header = ""
	resolver := config.Resolver

	main := resolver.NewPackage("main", "")
	runtimepkg := resolver.NewPackage("m/runtime", "")
	providerpkg := main

	runtimeModule, err := RuntimeModule(runtimepkg, resolver)
	if err != nil {
		t.Fatalf("unexpected error %+v", err)
	}
	providerModule, err := ProviderModule(providerpkg, resolver, "Provider")
	if err != nil {
		t.Fatalf("unexpected error %+v", err)
	}

	cases := []struct {
		msg      string
		here     *tinypkg.Package
		mount    func(r *web.Router)
		override func(t *resolve.Tracker)

		want   string
		hasErr bool
	}{
		{
			msg:   "no-deps",
			here:  main,
			mount: func(r *web.Router) { r.Get("/ping", Ping) },
			want: `package main

import (
	"github.com/podhmo/apikit/web/webgen/gen-chi"
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
	"github.com/podhmo/apikit/web/webgen/gen-chi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var pathParams struct {
			` + "message string `query:\"message,required\"`" + `
		}
		if err := runtime.BindPathParams(&pathParams, req, "message"); err != nil {
			w.WriteHeader(404)
			runtime.HandleResult(w, req, nil, err); return
		}
		result, err := genchi.Greeting(pathParams.message)
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},
		{
			msg:   "bind-path-with-new-type",
			here:  main,
			mount: func(r *web.Router) { r.Get("/greet/{message}", GreetingWithNewType) },
			want: `package main

import (
	"github.com/podhmo/apikit/web/webgen/gen-chi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var pathParams struct {
			` + "message genchi.Message `query:\"message,required\"`" + `
		}
		if err := runtime.BindPathParams(&pathParams, req, "message"); err != nil {
			w.WriteHeader(404)
			runtime.HandleResult(w, req, nil, err); return
		}
		result, err := genchi.GreetingWithNewType(pathParams.message)
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},
		{
			msg:   "bind-query",
			here:  main,
			mount: func(r *web.Router) { r.Get("/greet/{message}", GreetingWithQueryString) },
			want: `package main

import (
	"github.com/podhmo/apikit/web/webgen/gen-chi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var pathParams struct {
			` + "message genchi.Message `query:\"message,required\"`" + `
		}
		if err := runtime.BindPathParams(&pathParams, req, "message"); err != nil {
			w.WriteHeader(404)
			runtime.HandleResult(w, req, nil, err); return
		}
		var queryParams struct {
			` + "verbose *bool `query:\"verbose\"`" + `
		}
		if err := runtime.BindQuery(&queryParams, req); err != nil {
			_ = err // ignored
		}
		result, err := genchi.GreetingWithQueryString(pathParams.message, queryParams.verbose)
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},
		// TODO: ng 404
		// TODO: path validation
		{
			msg:   "bind-data",
			here:  main,
			mount: func(r *web.Router) { r.Post("/message", PostMessage) },
			want: `package main

import (
	"github.com/podhmo/apikit/web/webgen/gen-chi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var data genchi.Data
		if err := runtime.BindBody(&data, req.Body); err != nil {
			w.WriteHeader(400)
			runtime.HandleResult(w, req, nil, err); return
		}
		if err := runtime.ValidateStruct(&data); err != nil {
			w.WriteHeader(422)
			runtime.HandleResult(w, req, nil, err); return
		}
		result, err := genchi.PostMessage(data)
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},
		{
			msg:  "ng-bind-data",
			here: main,
			mount: func(r *web.Router) {
				r.Post("/message", func(data Data, data2 Data) (interface{}, error) { return nil, nil })
			},
			hasErr: true,
		},
		{
			msg:   "single-dep",
			here:  main,
			mount: func(r *web.Router) { r.Get("/articles", ListArticle) },
			want: `package main

import (
	"github.com/podhmo/apikit/web/webgen/gen-chi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err); return
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
				tracker.Override("db", func() (*DB, error) { return nil, nil })
			},
			want: `package main

import (
	"github.com/podhmo/apikit/web/webgen/gen-chi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err); return
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
	"github.com/podhmo/apikit/web/webgen/gen-chi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, _, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err); return
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
	"github.com/podhmo/apikit/web/webgen/gen-chi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err); return
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
		// loginRequired (external dependencies)
		{
			msg:  "single-dep-with-context-with-external-dep",
			here: main,
			mount: func(r *web.Router) {
				r.Get("/articles", ListArticleWithContext, web.WithExtraDependencies(LoginRequired))
			},
			want: `package main

import (
	"context"
	"github.com/podhmo/apikit/web/webgen/gen-chi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err); return
		}
		var ctx context.Context = req.Context()
		var db *genchi.DB
		{
			db = provider.DB()
		}
		if err := genchi.LoginRequired(db); err != nil {
			runtime.HandleResult(w, req, nil, err); return
		}
		result, err := genchi.ListArticleWithContext(ctx, db)
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},
		{
			msg:  "no-deps-with-external-dep",
			here: main,
			mount: func(r *web.Router) {
				r.Get("/ping", Ping, web.WithExtraDependencies(LoginRequired))
			},
			want: `package main

import (
	"github.com/podhmo/apikit/web/webgen/gen-chi"
	"net/http"
	"m/runtime"
)

func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err); return
		}
		var db *genchi.DB
		{
			db = provider.DB()
		}
		if err := genchi.LoginRequired(db); err != nil {
			runtime.HandleResult(w, req, nil, err); return
		}
		result, err := genchi.Ping()
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			translator := &Translator{
				Resolver:       config.Resolver,
				Tracker:        resolve.NewTracker(config.Resolver),
				Config:         config.Config,
				RuntimeModule:  runtimeModule,
				ProviderModule: providerModule,
			}

			r := web.NewRouter()
			c.mount(r)
			if c.override != nil {
				c.override(translator.Tracker)
			}

			if err := web.Walk(r, func(n *web.WalkerNode) error {
				code := translator.TranslateToHandler(c.here, n, handlerName)
				var buf strings.Builder
				err := code.Emit(&buf)

				if c.hasErr {
					if err == nil {
						t.Error("expected error, but not occured")
					}
					return nil
				}

				if err != nil {
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
