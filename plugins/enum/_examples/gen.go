//go:build apikit
// +build apikit

package main

import (
	"fmt"

	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/plugins"
	"github.com/podhmo/apikit/plugins/enum"
)

func main() {
	emitgo.NewConfigFromRelativePath(main, ".").MustEmitWith(func(emitter *emitgo.Emitter) error {
		pc := plugins.NewDefaultPluginContext(emitter)
		pkg := emitter.RootPkg.Relative("generated", "")

		enumset := enum.StringEnums("Grade", "s", "a", "b", "c", "d")
		if err := pc.IncludePlugin(pkg, enum.Options{EnumSet: enumset}); err != nil {
			return fmt.Errorf("generate Grade: %w", err)
		}
		return nil
	})
}
