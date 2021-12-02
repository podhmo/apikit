//go:build apikitseed
// +build apikitseed

package main

import (
	"fmt"

	"m/13openapi/seed"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/plugins"
	"github.com/podhmo/apikit/plugins/enum"
)

func main() {
	emitgo.NewConfigFromRelativePath(main, ".").MustEmitWith(func(emitter *emitgo.Emitter) error {
		emitter.DisableManagement = true

		pc := plugins.NewDefaultPluginContext(emitter)
		pkg := emitter.RootPkg.Relative("design/enum", "")

		enumset := seed.Enums.SortOrder
		if err := pc.IncludePlugin(pkg, enum.Options{EnumSet: enumset}); err != nil {
			return fmt.Errorf("generate Grade: %w", err)
		}
		return nil
	})
}
