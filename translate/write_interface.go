package translate

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/tinypkg"
)

func collectImports(here *tinypkg.Package, t *Tracker) ([]*tinypkg.ImportedPackage, error) {
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
		imports = append(imports, here.Import(sym.Package))
	}
	for _, need := range t.Needs {
		if err := tinypkg.Walk(need.raw, use); err != nil {
			return err
		}
	}
	return imports
}

// TODO: import
// TODO: same package
func WriteInterface(w io.Writer, here *tinypkg.Package, t *Tracker, name string) {
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
		typeExpr := resolve.ExtractSymbol(here, need.raw.Shape).String()
		fmt.Fprintf(w, "\t%s() %s\n", methodName, typeExpr)
		if _, duplicated := usedNames[methodName]; duplicated {
			log.Printf("WARN: method name %s is duplicated", methodName)
		}
		usedNames[methodName] = true
	}
	io.WriteString(w, "}\n")
}
