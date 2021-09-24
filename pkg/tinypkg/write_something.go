package tinypkg

import (
	"fmt"
	"io"
	"strings"
)

func WriteFunc(w io.Writer, here *Package, name string, f *Func, body func() error) error {
	if name == "" {
		name = f.Name
	}

	args := make([]string, 0, len(f.Args))
	for _, x := range f.Args {
		args = append(args, ToRelativeTypeString(here, x))
	}

	returns := make([]string, 0, len(f.Returns))
	for _, x := range f.Returns {
		returns = append(returns, ToRelativeTypeString(here, x))
	}

	// func <name>(<args>...) (<returns>) {
	// ...
	// }
	defer fmt.Fprintln(w, "}")
	switch len(returns) {
	case 0:
		fmt.Fprintf(w, "func %s(%s) {\n", name, strings.Join(args, ", "))
	case 1:
		fmt.Fprintf(w, "func %s(%s) %s {\n", name, strings.Join(args, ", "), returns[0])
	default:
		fmt.Fprintf(w, "func %s(%s) (%s) {\n", name, strings.Join(args, ", "), strings.Join(returns, ", "))
	}
	return body()
}

func WriteInterface(w io.Writer, here *Package, name string, iface *Interface) error {
	if name == "" {
		name = iface.Name
	}
	// interface <name> {
	// ..
	//}
	fmt.Fprintf(w, "type %s interface {\n", name)
	defer fmt.Fprintln(w, "}")
	for _, method := range iface.Methods {
		fmt.Fprintf(w, "\t%s\n", ToInterfaceMethodString(here, method.Name, method))
	}
	return nil
}

type Binding struct {
	Indent string

	Name     string
	Provider *Func

	HasError   bool
	HasCleanup bool
}

// TODO: support not-pointer zero value

func (b *Binding) WriteWithCallbackAndError(w io.Writer, here *Package, returns []*Var) error {
	if len(returns) == 0 {
		return fmt.Errorf("invalid number of returns %+v", returns)
	}

	fmt.Fprintf(w, "%svar %s %s\n", b.Indent, b.Name, ToRelativeTypeString(here, returns[0]))
	fmt.Fprintf(w, "%s{\n", b.Indent)
	defer fmt.Fprintf(w, "%s}\n", b.Indent)
	{
		if b.HasCleanup {
			fmt.Fprintf(w, "%s\tvar cleanup func()\n", b.Indent)
		}
		if b.HasError {
			fmt.Fprintf(w, "%s\tvar err error\n", b.Indent)
		}

		var callRHS string
		{
			provider := b.Provider
			args := make([]string, 0, len(provider.Args))
			for _, x := range provider.Args {
				args = append(args, x.Name)
			}
			callRHS = fmt.Sprintf("%s(%s)", provider.Name, strings.Join(args, ", "))
			if provider.Recv != "" {
				callRHS = fmt.Sprintf("%s.%s", provider.Recv, callRHS)
			}
		}

		switch len(returns) {
		case 1:
			fmt.Fprintf(w, "%s\t%s = %s\n", b.Indent, b.Name, callRHS)
		case 2:
			if b.HasError {
				fmt.Fprintf(w, "%s\t%s, err = %s\n", b.Indent, b.Name, callRHS)
			} else if b.HasCleanup {
				fmt.Fprintf(w, "%s\t%s, cleanup = %s\n", b.Indent, b.Name, callRHS)
			}
		case 3:
			fmt.Fprintf(w, "%s\t%s, cleanup, err = %s\n", b.Indent, b.Name, callRHS)
		}

		if b.HasCleanup {
			fmt.Fprintf(w, "%s\tif cleanup != nil {\n", b.Indent)
			fmt.Fprintf(w, "%s\t\tdefer cleanup()\n", b.Indent)
			fmt.Fprintf(w, "%s\t}\n", b.Indent)
		}
		if b.HasError { // TODO: support zero-value
			var returnRHS string
			{
				values := []string{"nil", "nil", "nil"}
				if returns[len(returns)-1].Node.String() == "error" {
					values[len(returns)-1] = "err"
				}
				returnRHS = strings.Join(values[:len(returns)], ", ")
			}

			fmt.Fprintf(w, "%s\tif err != nil {\n", b.Indent)
			fmt.Fprintf(w, "%s\t\treturn %s\n", b.Indent, returnRHS)
			fmt.Fprintf(w, "%s\t}\n", b.Indent)
		}
	}
	return nil
}
