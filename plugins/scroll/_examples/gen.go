//go:build apikit
// +build apikit

package main

import (
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/plugins"
	"github.com/podhmo/apikit/plugins/scroll"
)

func main() {
	emitgo.NewConfigFromRelativePath(main, ".").MustEmitWith(func(emitter *emitgo.Emitter) error {
		pc := plugins.NewDefaultPluginContext(emitter)
		pkg := emitter.RootPkg.Relative("runtime", "")

		var latestID int = 0 // for scroll implemention
		return pc.IncludePlugin(pkg, scroll.Options{LatestIDTypeZeroValue: latestID})
	})
}
