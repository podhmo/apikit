// +build apikit

package main

import (
	"fmt"
	"log"

	"github.com/podhmo/apikit/ext"
	"github.com/podhmo/apikit/ext/enum"
	"github.com/podhmo/apikit/pkg/emitgo"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() (err error) {
	emitter := emitgo.NewConfigFromRelativePath(main, ".").NewEmitter()
	defer emitter.EmitWith(&err)

	pc := ext.NewDefaultPluginContext(emitter)
	pkg := emitter.RootPkg.Relative("generated", "")

	if err := pc.IncludePlugin(pkg, enum.StringEnums("Grade", "S", "A", "B", "C", "D")); err != nil {
		return fmt.Errorf("generate Grade: %w", err)
	}
	return nil
}
