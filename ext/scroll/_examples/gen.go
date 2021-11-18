// +build apikit

package main

import (
	"github.com/podhmo/apikit/ext"
	"github.com/podhmo/apikit/ext/scroll"
	"github.com/podhmo/apikit/pkg/emitgo"
)

func main() {
	emitgo.NewConfigFromRelativePath(main, ".").MustEmitWith(func(emitter *emitgo.Emitter) error {
		pc := ext.NewDefaultPluginContext(emitter)
		pkg := emitter.RootPkg.Relative("runtime", "")

		var latestID int = 0 // for scroll implemention
		return pc.IncludePlugin(pkg, scroll.Options{LatestIDTypeZeroValue: latestID})
	})
}
