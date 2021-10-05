package webtranslate

import (
	"io"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

func (t *Translator) TranslateToInterface(here *tinypkg.Package) *code.Code {
	name := t.Config.ProviderName // xxx
	return &code.Code{
		Name: name,
		Here: here,
		// priority: code.PriorityFirst,
		Config: t.Config.Config,
		ImportPackages: func() ([]*tinypkg.ImportedPackage, error) {
			return collectImportsForInterface(here, t.Resolver, t.Tracker)
		},
		EmitCode: func(w io.Writer) error {
			return writeInterface(w, here, t.Resolver, t.Tracker, name)
		},
	}
}

func collectImportsForInterface(here *tinypkg.Package, resolver *resolve.Resolver, t *resolve.Tracker) ([]*tinypkg.ImportedPackage, error) {
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

func writeInterface(w io.Writer, here *tinypkg.Package, resolver *resolve.Resolver, t *resolve.Tracker, name string) error {
	iface := t.ExtractInterface(here, resolver, name)
	return tinypkg.WriteInterface(w, here, name, iface)
}