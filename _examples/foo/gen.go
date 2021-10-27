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
	"m/foo/design"
)

// generate code: VERBOSE=1 go run gen.go

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() (err error) {
	emitter := emitgo.NewConfigFromRelativePath(action.Hello, "..").NewEmitter()
	emitter.FilenamePrefix = "gen_" // generated file name is "gen_<name>.go"
	defer emitter.EmitWith(&err)

	r := web.NewRouter()
	r.Get("/hello", action.Hello)

	c := genchi.DefaultConfig()
	// c.Override("logger", action.NewLogger) // register provider as func() (*log.Logger, error)

	g := c.New(emitter)
	if err := g.Generate(context.Background(), r, design.HTTPStatusOf); err != nil {
		return err
	}

	// use scroll plugin (string type version)
	// g.IncludePlugin(g.RuntimePkg, scroll.Options{LatestIDTypeZeroValue: ""}) // latestId is string
	return nil
}
