// +build apikit

package main

import (
	"context"
	"log"

	"m/14lambda/design"

	"github.com/podhmo/apikit/pkg/emitgo"
	genlambda "github.com/podhmo/apikit/web/webgen/gen-lambda"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() (err error) {
	emitter := emitgo.NewConfigFromRelativePath(run, ".").NewEmitter()
	emitter.FilenamePrefix = "gen_" // generated file name is "gen_<name>.go"
	defer emitter.EmitWith(&err)

	ctx := context.Background()

	c := genlambda.DefaultConfig()
	c.Header = "" // todo: remove

	g := c.New(emitter)

	{
		pkg := emitter.RootPkg.Relative("lambdahandler/postarticlecomment", "")
		action := design.PostArticleComment
		if err := g.Generate(ctx, pkg, action); err != nil {
			return err
		}
	}

	return nil
}
