// +build apikit

package main

import (
	"log"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/web/ext"
	"github.com/podhmo/apikit/web/ext/scroll"
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
	pkg := emitter.RootPkg.Relative("runtime", "")

	var latestID int = 0 // for scroll implemention
	pc.IncludePlugin(pkg, scroll.Options{LatestIDTypeZeroValue: latestID})
	return nil
}
