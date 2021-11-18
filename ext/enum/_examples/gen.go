// +build apikit

package main

import (
	"fmt"

	"github.com/podhmo/apikit/ext"
	"github.com/podhmo/apikit/ext/enum"
	"github.com/podhmo/apikit/pkg/emitgo"
)

func main() {
	c := emitgo.NewConfigFromRelativePath(main, ".")
	c.MustEmitWith(func(emitter *emitgo.Emitter) error {
		pc := ext.NewDefaultPluginContext(emitter)
		pkg := emitter.RootPkg.Relative("generated", "")

		enumset := enum.StringEnums("Grade", "s", "a", "b", "c", "d")
		if err := pc.IncludePlugin(pkg, enum.Options{EnumSet: enumset}); err != nil {
			return fmt.Errorf("generate Grade: %w", err)
		}
		return nil
	})
}
