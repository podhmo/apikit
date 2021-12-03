//go:build apikit
// +build apikit

package main

import (
	"context"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/plugins"
	"github.com/podhmo/apikit/plugins/scroll"
)

func main() {
	emitgo.NewConfigFromRelativePath(main, ".").MustEmitWith(func(emitter *emitgo.Emitter) error {
		ctx := context.Background()
		pkg := emitter.RootPkg.Relative("runtime", "")

		var latestID int = 0 // for scroll implemention
		return plugins.NewDefaultConfig(emitter).ActivatePlugins(ctx, pkg,
			scroll.Options{LatestIDTypeZeroValue: latestID},
		)
	})
}
