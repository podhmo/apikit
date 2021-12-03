//go:build apikitseed
// +build apikitseed

package main

import (
	"context"
	"m/13openapi/seed"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/plugins"
	"github.com/podhmo/apikit/plugins/enum"
)

func main() {
	emitgo.NewConfigFromRelativePath(main, ".").MustEmitWith(func(emitter *emitgo.Emitter) error {
		emitter.DisableManagement = true

		ctx := context.Background()
		pkg := emitter.RootPkg.Relative("design/enum", "")

		c := plugins.NewDefaultConfig(emitter)
		return c.ActivatePlugins(ctx, pkg,
			enum.Options{EnumSet: seed.Enums.SortOrder},
		)
	})
}
