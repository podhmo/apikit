package webtranslate

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

	translator := NewTranslator(DefaultConfig())
	resolver := translator.Resolver

	main := resolver.NewPackage("main", "")
	runtime := resolver.NewPackage("m/runtime", "")

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
			want: `
func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) {
	return func(w http.ResponseWriter, req *http.Request) http.HandlerFunc{
		result, err := webtranslate.Ping()
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},
		{
			msg:   "bind-path",
			here:  main,
			mount: func(r *web.Router) { r.Get("/greet/{message}", Greeting) },
			want: `
func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) {
	return func(w http.ResponseWriter, req *http.Request) http.HandlerFunc{
		message := runtime.PathParam(req, "message")
		result, err := webtranslate.Greeting(message)
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
			want: `
func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) {
	return func(w http.ResponseWriter, req *http.Request) http.HandlerFunc{
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err)
			return
		}
		var db *webtranslate.DB
		{
			db = provider.DB()
		}
		result, err := webtranslate.ListArticle(db)
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
			want: `
func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) {
	return func(w http.ResponseWriter, req *http.Request) http.HandlerFunc{
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err)
			return
		}
		var db *webtranslate.DB
		{
			var err error
			db, err = provider.DB()
			if err != nil {
				runtime.HandleResult(w, req, nil, err); return
			}
		}
		result, err := webtranslate.ListArticle(db)
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
			want: `
func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) {
	return func(w http.ResponseWriter, req *http.Request) http.HandlerFunc{
		req, _, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err)
			return
		}
		ctx := req.Context()
		result, err := webtranslate.PingWithContext(ctx)
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},
		{
			msg:   "single-dep-with-context",
			here:  main,
			mount: func(r *web.Router) { r.Get("/articles", ListArticleWithContext) },
			want: `
func Handler(getProvider func(*http.Request) (*http.Request, Provider, error)) {
	return func(w http.ResponseWriter, req *http.Request) http.HandlerFunc{
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err)
			return
		}
		ctx := req.Context()
		var db *webtranslate.DB
		{
			db = provider.DB()
		}
		result, err := webtranslate.ListArticleWithContext(ctx, db)
		runtime.HandleResult(w, req, result, err)
	}
}`,
		},

		// TODO: unexpected action
		// TODO: path binding
		// TODO: handling error
	}
	// ?? runtime is also interface?

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			r := web.NewRouter()
			c.mount(r)

			var node *web.WalkerNode
			web.Walk(r, func(n *web.WalkerNode) error {
				node = n
				return nil
			})

			tracker := resolve.NewTracker()
			if c.override != nil {
				c.override(tracker)
			}

			providerModule, err := translator.GetProviderModule(runtime, "Provider")
			if err != nil {
				t.Fatalf("unexpected error %+v", err)
			}
			runtimeModule, err := translator.RuntimeModule(runtime)
			if err != nil {
				t.Fatalf("unexpected error %+v", err)
			}

			def := resolver.Def(node.Node.Value)
			tracker.Track(def)
			pathinfo, err := web.ExtractPathInfo(node.Node.VariableNames, def)
			if err != nil {
				t.Fatalf("unexpected error, extract info, %+v", err)
			}

			var buf strings.Builder
			if err := WriteHandlerFunc(&buf, c.here, resolver, tracker, pathinfo, providerModule, runtimeModule, handlerName); err != nil {
				t.Errorf("unexpected error %+v", err)
			}
			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(buf.String()); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}
