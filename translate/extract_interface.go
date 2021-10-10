package translate

import (
	"io"
	"strings"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/tinypkg"
)

func (t *Translator) ExtractProviderInterface(here *tinypkg.Package, name string) *code.CodeEmitter {
	t.providerVar = &tinypkg.Var{Name: strings.ToLower(name), Node: here.NewSymbol(name)}
	c := &code.Code{
		Name: name,
		Here: here,
		// priority: code.PriorityFirst,
		Config: t.Config,
		ImportPackages: func(collector *tinypkg.ImportCollector) error {
			resolver := t.Resolver
			use := collector.Collect
			here := collector.Here
			for _, need := range t.Needs {
				shape := need.Shape
				if need.OverrideDef != nil {
					shape = need.OverrideDef.Shape
				}
				sym := resolver.Symbol(here, shape)
				if err := tinypkg.Walk(sym, use); err != nil {
					return err
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
