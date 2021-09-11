package translate

import (
	"fmt"
	"io"
	"strings"

	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/tinypkg"
)

// TODO: omit provider arguments

func (t *Translator) TranslateToRunner(here *tinypkg.Package, def *resolve.Def, name string, provider *tinypkg.Var) *Code {
	return &Code{
		Name:     name,
		Here:     here,
		EmitFunc: t.EmitFunc,
		ImportPackages: func() ([]*tinypkg.ImportedPackage, error) {
			return collectImports(here, t.Tracker)
		},
		EmitCode: func(w io.Writer) error {
			return writeRunner(w, here, def, name, provider)
		},
	}
}

func writeRunner(w io.Writer, here *tinypkg.Package, def *resolve.Def, name string, provider *tinypkg.Var) error {
	// TODO:
	// get components
	// rest as arguments
	// TODO: components
	// TODO: handling context.Context

	var components []resolve.Item
	var ignored []string
	argNames := make([]string, 0, len(def.Args))
	args := make([]string, 0, len(def.Args)+1)
	{
		args = append(args, tinypkg.ToRelativeTypeString(here, provider))
	}
	for _, x := range def.Args {
		argNames = append(argNames, x.Name)

		if x.Kind == resolve.KindComponent {
			components = append(components, x)
			continue
		}

		sym := resolve.ExtractSymbol(here, x.Shape)
		if x.Kind == resolve.KindIgnored { // e.g. context.Context
			ignored = append(ignored, fmt.Sprintf("%s %s", x.Name, sym.String()))
		} else {
			args = append(args, fmt.Sprintf("%s %s", x.Name, sym.String()))
		}
	}
	if len(ignored) > 0 {
		args = append(ignored, args...)
	}

	returns := make([]string, 0, len(def.Returns))
	for _, x := range def.Returns {
		sym := resolve.ExtractSymbol(here, x.Shape)
		returns = append(returns, sym.String()) // TODO: need using x.Name?
	}

	// func <name>(<args>...) (<returns>) {
	switch len(returns) {
	case 0:
		fmt.Fprintf(w, "func %s(%s) {\n", name, strings.Join(args, ", "))
	case 1:
		fmt.Fprintf(w, "func %s(%s) %s {\n", name, strings.Join(args, ", "), returns[0])
	default:
		fmt.Fprintf(w, "func %s(%s) (%s) {\n", name, strings.Join(args, ", "), strings.Join(returns, ", "))
	}

	if len(components) > 0 {
		for _, x := range components {
			fmt.Fprintln(w, "\t{")
			sym := resolve.ExtractSymbol(here, x.Shape)
			fmt.Fprintf(w, "\t\tvar %s %s\n", x.Name, tinypkg.ToRelativeTypeString(here, sym))

			// TODO: communicate with write_interface.go's functions
			methodName := x.Shape.GetReflectType().Name()
			fmt.Fprintf(w, "\t\t%s = %s.%s()\n", x.Name, provider.Name, methodName)
			fmt.Fprintln(w, "\t}")
		}
	}

	// return <inner function>(<args>...)
	fmt.Fprintf(w, "\treturn %s(%s)\n",
		tinypkg.ToRelativeTypeString(here, def.Symbol),
		strings.Join(argNames, ", "),
	)

	// }
	fmt.Fprintln(w, "}")
	return nil
}
