//go:build apikit
// +build apikit

// this code is generated by "apikit init"

package main

import (
	"context"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/web"
	"github.com/podhmo/apikit/web/webgen/gen-chi"
	"log"
	"m/foo/action"
	"m/foo/design/code"
)

// generate code: VERBOSE=1 go run gen.go

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func newRouter() *web.Router {
	r := web.NewRouter()
	r.Get("/hello", action.Hello)
	return r
}

func run() error {
	ctx := context.Background()
	return emitgo.NewConfigFromRelativePath(action.Hello, "..").EmitWith(func(emitter *emitgo.Emitter) error {
		emitter.FilenamePrefix = "gen_" // generated file name is "gen_<name>.go"

		c := genchi.DefaultConfig()
		// c.Override("logger", action.NewLogger) // register provider as func() (*log.Logger, error)

		g := c.New(emitter)
		r := newRouter()
		if err := g.Generate(ctx, r, code.HTTPStatusOf); err != nil {
			return err
		}

		// // use scroll plugin (string type version)
		// return g.ActivatePlugins(ctx, g.RuntimePkg,
		// 	scroll.Options{LatestIDTypeZeroValue: ""}, // latestId is string
		// )
		return nil
	})
}
