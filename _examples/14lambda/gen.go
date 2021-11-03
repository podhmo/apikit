// +build apikit

package main

import (
	"context"
	"log"

	"m/14lambda/design"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
)

type Generator interface {
	Generate(ctx context.Context, pkg *tinypkg.Package, action interface{}) error
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() (err error) {
	emitter := emitgo.NewConfigFromRelativePath(run, "").NewEmitter()
	emitter.FilenamePrefix = "gen_" // generated file name is "gen_<name>.go"
	defer emitter.EmitWith(&err)

	ctx := context.Background()
	var g Generator

	{
		pkg := emitter.RootPkg.Relative("lambdahandler/postarticlecomment", "")
		action := design.PostArticleComment
		if err := g.Generate(ctx, pkg, action); err != nil {
			return err
		}
	}

	return nil
}
