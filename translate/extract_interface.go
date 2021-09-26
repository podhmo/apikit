package translate

import (
	"io"
	"strings"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

func (t *Translator) ExtractInterface(here *tinypkg.Package, name string) *code.Code {
	t.providerVar = &tinypkg.Var{Name: strings.ToLower(name), Node: here.NewSymbol(name)}
	return &code.Code{
		Name: name,
		Here: here,
		// priority: code.PriorityFirst,
		Config: t.Config,
		ImportPackages: func() ([]*tinypkg.ImportedPackage, error) {
			return collectImportsForProviderInterface(here, t.Resolver, t.Tracker)
		},
		EmitCode: func(w io.Writer) error {
			return writeProviderInterface(w, here, t.Resolver, t.Tracker, name)
		},
	}
}

func collectImportsForProviderInterface(here *tinypkg.Package, resolver *resolve.Resolver, t *resolve.Tracker) ([]*tinypkg.ImportedPackage, error) {
	collector := tinypkg.NewImportCollector(here)
	use := collector.Collect
	for _, need := range t.Needs {
		shape := need.Shape
		if need.OverrideDef != nil {
			shape = need.OverrideDef.Shape
		}
		sym := resolver.Symbol(here, shape)
		if err := tinypkg.Walk(sym, use); err != nil {
			return nil, err
		}
	}
	return collector.Imports, nil
}

func writeProviderInterface(w io.Writer, here *tinypkg.Package, resolver *resolve.Resolver, t *resolve.Tracker, name string) error {
	iface := t.ExtractInterface(here, resolver, name)
	return tinypkg.WriteInterface(w, here, name, iface)
}
