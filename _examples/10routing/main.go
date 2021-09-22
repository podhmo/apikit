package main

import (
	"context"
	"log"

	"m/10routing/design"

	"github.com/podhmo/apikit"
	"github.com/podhmo/apikit/pkg/emitgo"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() (err error) {
	emitter := emitgo.NewFromRelativePath(design.Hello, "..")
	defer emitter.EmitWith(&err)

	here := emitter.RootPkg.Relative("router", "")
	{
		code := apikit.GenerateRouterCode(here)
		emitter.Register(here, "router.go", code)
	}
	{
		var fn design.HandlerFunc = func(ctx context.Context) (interface{}, error) { return nil, nil }
		_, code := apikit.GenerateTypeCode(here, fn)
		emitter.Register(here, "types.go", code)
	}
	return nil
}
