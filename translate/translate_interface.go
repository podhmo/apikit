package translate

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/tinypkg"
)

func (t *Translator) TranslateToInterface(here *tinypkg.Package, name string) *Code {
	return &Code{
		Name:     name,
		Here:     here,
		EmitFunc: t.EmitFunc,
		ImportPackages: func() ([]*tinypkg.ImportedPackage, error) {
			return collectImportsForInterface(here, t.Tracker)
		},
		EmitCode: func(w io.Writer) error {
			return writeInterface(w, here, t.Tracker, name)
		},
	}
}

func collectImportsForInterface(here *tinypkg.Package, t *Tracker) ([]*tinypkg.ImportedPackage, error) {
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
		sym := resolve.ExtractSymbol(here, need.Shape)
		if err := tinypkg.Walk(sym, use); err != nil {
			return nil, err
		}
	}
	return imports, nil
}

func writeInterface(w io.Writer, here *tinypkg.Package, t *Tracker, name string) error {
	fmt.Fprintf(w, "type %s interface {\n", name)
	usedNames := map[string]bool{}
	for _, need := range t.Needs {
		k := need.rt
		methodName := need.rt.Name()
		if len(t.seen[k]) > 1 {
			methodName = strings.ToUpper(string(need.Name[0])) + need.Name[1:] // TODO: use GoName
		}

		// TODO: T, (T, error)
		// TODO: support correct type expression
		typeExpr := resolve.ExtractSymbol(here, need.Shape).String()
		fmt.Fprintf(w, "\t%s() %s\n", methodName, typeExpr)
		if _, duplicated := usedNames[methodName]; duplicated {
			log.Printf("WARN: method name %s is duplicated", methodName)
		}
		usedNames[methodName] = true
	}
	io.WriteString(w, "}\n")
	return nil
}
