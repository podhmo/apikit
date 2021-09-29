package webtranslate

import (
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

func ListArticle(db *DB) ([]*Article, error) {
	return nil, nil
}

func TestWriteHandlerFunc(t *testing.T) {
	handlerName := "Handler"

	config := DefaultConfig()
	resolver := config.Resolver

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
				tracker.Override(reflect.TypeOf(&DB{}), "DBOrError", resolver.Def(func() (*DB, error) { return nil, nil }))
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
				runtime.HandleResult(w, req, nil, err)
				return
		}
		result, err := webtranslate.ListArticle(db)
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

			def := resolver.Def(node.Node.Value)
			tracker.Track(def)
			pathinfo, err := web.ExtractPathInfo(node.Node.VariableNames, def)
			if err != nil {
				t.Fatalf("unexpected error, extract info, %+v", err)
			}

			var buf strings.Builder
			providerFunc := c.here.NewFunc( // todo: simplify
				"getProvider",
				[]*tinypkg.Var{{Node: &tinypkg.Pointer{Lv: 1, V: resolve.NewResolver().NewPackage("net/http", "").NewSymbol("Request")}}},
				[]*tinypkg.Var{
					{Node: &tinypkg.Pointer{Lv: 1, V: resolver.NewPackage("net/http", "").NewSymbol("Request")}},
					{Node: c.here.NewSymbol("Provider")},
					{Node: tinypkg.NewSymbol("error")},
				},
			)
			if err := WriteHandlerFunc(&buf, c.here, resolver, tracker, pathinfo, runtime, providerFunc, handlerName); err != nil {
				t.Errorf("unexpected error %+v", err)
			}
			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(buf.String()); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}
