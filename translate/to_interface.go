package translate

import (
	"io"
	"strings"

	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

func (t *Translator) TranslateToInterface(here *tinypkg.Package, name string) *Code {
	t.providerVar = &tinypkg.Var{Name: strings.ToLower(name), Node: here.NewSymbol(name)}

	return &Code{
		Name:     name,
		Here:     here,
		priority: priorityFirst,
		emitFunc: t.EmitFunc,
		ImportPackages: func() ([]*tinypkg.ImportedPackage, error) {
			return collectImportsForInterface(here, t.Resolver, t.Tracker)
		},
		EmitCode: func(w io.Writer) error {
			return writeInterface(w, here, t.Resolver, t.Tracker, name)
		},
	}
}

func collectImportsForInterface(here *tinypkg.Package, resolver *resolve.Resolver, t *Tracker) ([]*tinypkg.ImportedPackage, error) {
	collector := tinypkg.NewImportCollector(here)
	use := collector.Collect
	for _, need := range t.Needs {
		shape := need.Shape
		if need.overrideDef != nil {
			shape = need.overrideDef.Shape
		}
		sym := resolver.Symbol(here, shape)
		if err := tinypkg.Walk(sym, use); err != nil {
			return nil, err
		}
	}
	return collector.Imports, nil
}

func writeInterface(w io.Writer, here *tinypkg.Package, resolver *resolve.Resolver, t *Tracker, name string) error {
	iface := t.extractInterface(here, resolver, name)
	return tinypkg.WriteInterface(w, here, name, iface)
}
