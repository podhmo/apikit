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
	if name == "" {
		name = def.Name
	}
	if provider == nil {
		provider = t.providerVar
	}

	return &Code{
		Name:     name,
		Here:     here,
		EmitFunc: t.EmitFunc,
		ImportPackages: func() ([]*tinypkg.ImportedPackage, error) {
			return collectImportsForRunner(here, t.Resolver, def, provider)
		},
		EmitCode: func(w io.Writer) error {
			return writeRunner(w, here, t.Resolver, def, provider, name)
		},
	}
}

func collectImportsForRunner(here *tinypkg.Package, resolver *resolve.Resolver, def *resolve.Def, provider *tinypkg.Var) ([]*tinypkg.ImportedPackage, error) {
	imports := make([]*tinypkg.ImportedPackage, 0, len(def.Args)+len(def.Returns))
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

	for _, x := range def.Args {
		sym := resolver.Symbol(here, x.Shape)
		if err := tinypkg.Walk(sym, use); err != nil {
			return nil, err
		}
	}
	for _, x := range def.Returns {
		sym := resolver.Symbol(here, x.Shape)
		if err := tinypkg.Walk(sym, use); err != nil {
			return nil, err
		}
	}
	if err := tinypkg.Walk(provider, use); err != nil {
		return nil, err
	}
	return imports, nil
}

func writeRunner(w io.Writer, here *tinypkg.Package, resolver *resolve.Resolver, def *resolve.Def, provider *tinypkg.Var, name string) error {
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

		sym := resolver.Symbol(here, x.Shape)
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
		sym := resolver.Symbol(here, x.Shape)
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

	// var <component> <type>
	// {
	//   <component> = <provider>.<method>()
	// }
	if len(components) > 0 {
		for _, x := range components {
			// TODO: communicate with write_interface.go's functions
			sym := resolver.Symbol(here, x.Shape)
			methodName := x.Shape.GetReflectType().Name()

			fmt.Fprintf(w, "\tvar %s %s\n", x.Name, tinypkg.ToRelativeTypeString(here, sym))
			fmt.Fprintln(w, "\t{")
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
