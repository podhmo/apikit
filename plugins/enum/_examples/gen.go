//go:build apikit
// +build apikit

package main

import (
	"context"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/plugins"
	"github.com/podhmo/apikit/plugins/enum"
)

func main() {
	emitgo.NewConfigFromRelativePath(main, ".").MustEmitWith(func(emitter *emitgo.Emitter) error {
		ctx := context.Background()
		here := emitter.RootPkg.Relative("generated", "")

		enumset := enum.StringEnums("Grade", "s", "a", "b", "c", "d")
		return plugins.NewDefaultConfig(emitter).ActivatePlugins(ctx, here,
			enum.Options{EnumSet: enumset},
		)
	})
}
