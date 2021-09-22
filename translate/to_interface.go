package translate

import (
	"fmt"
	"io"
	"log"
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
		EmitFunc: t.EmitFunc,
		ImportPackages: func() ([]*tinypkg.ImportedPackage, error) {
			return collectImportsForInterface(here, t.Resolver, t.Tracker)
		},
		EmitCode: func(w io.Writer) error {
			return writeInterface(w, here, t.Resolver, t.Tracker, name)
		},
	}
}

func collectImportsForInterface(here *tinypkg.Package, resolver *resolve.Resolver, t *Tracker) ([]*tinypkg.ImportedPackage, error) {
	imports := make([]*tinypkg.ImportedPackage, 0, len(t.Needs))
	seen := map[*tinypkg.Package]bool{}
	use := func(sym *tinypkg.Symbol) error {
		if sym.Package.Path == "" {
			return nil // bultins type (e.g. string, bool, ...)
		}
		if _, ok := seen[sym.Package]; ok {
			return nil
		}
		seen[sym.Package] = true
		if here == sym.Package {
			return nil
		}
		imports = append(imports, here.Import(sym.Package))
		return nil
	}
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
	return imports, nil
}

func writeInterface(w io.Writer, here *tinypkg.Package, resolver *resolve.Resolver, t *Tracker, name string) error {
	fmt.Fprintf(w, "type %s interface {\n", name)
	defer io.WriteString(w, "}\n")

	usedNames := map[string]bool{}
	for _, need := range t.Needs {
		k := need.rt

		methodName := need.rt.Name()
		if len(t.seen[k]) > 1 {
			methodName = strings.ToUpper(string(need.Name[0])) + need.Name[1:] // TODO: use GoName
		}
		shape := need.Shape
		if need.overrideDef != nil {
			shape = need.overrideDef.Shape
		}

		methodExpr := tinypkg.ToInterfaceMethodString(here, methodName, resolver.Symbol(here, shape))
		fmt.Fprintf(w, "\t%s\n", methodExpr)
		if _, duplicated := usedNames[methodName]; duplicated {
			log.Printf("WARN: method name %s is duplicated", methodName)
		}
		usedNames[methodName] = true
	}
	return nil
}
