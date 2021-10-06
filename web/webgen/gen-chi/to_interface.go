package genchi

import (
	"fmt"
	"io"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/tinypkg"
)

func (t *Translator) TranslateToInterface(here *tinypkg.Package, name string) *code.CodeEmitter {
	c := &code.Code{
		Name: name,
		Here: here,
		// priority: code.PriorityFirst,
		Config: t.Config,
		ImportPackages: func(collector *tinypkg.ImportCollector) error {
			resolver := t.Resolver
			tracker := t.Tracker
			here := collector.Here
			use := collector.Collect
			for _, need := range tracker.Needs {
				shape := need.Shape
				if need.OverrideDef != nil {
					shape = need.OverrideDef.Shape
				}
				sym := resolver.Symbol(here, shape)
				if err := tinypkg.Walk(sym, use); err != nil {
					return fmt.Errorf("on walk %s: %w", sym, err)
				}
			}
			return nil
		},
		EmitCode: func(w io.Writer, c *code.Code) error {
			iface := t.Tracker.ExtractInterface(here, t.Resolver, name)
			return tinypkg.WriteInterface(w, here, name, iface)
		},
	}
	return &code.CodeEmitter{Code: c}
}
