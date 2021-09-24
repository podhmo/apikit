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

var ErrUnexpectedReturnType = fmt.Errorf("unexpected-return-type")
var ErrUnexpectedExternalReturnType = fmt.Errorf("unexpected-external-return-type")

// TODO: support not-pointer zero value

func (b *Binding) WriteWithCallbackAndError(w io.Writer, here *Package, returns []*Var) error {
	if 3 < len(returns) {
		return fmt.Errorf("sorry the maximum value of supported number-of-return-value is 3, but %s is passed, %w", returns, ErrUnexpectedExternalReturnType)
	}

	fmt.Fprintf(w, "%svar %s %s\n", b.Indent, b.Name, ToRelativeTypeString(here, b.Provider.Returns[0].Node))
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

		switch len(b.Provider.Returns) {
		case 1:
			fmt.Fprintf(w, "%s\t%s = %s\n", b.Indent, b.Name, callRHS)
		case 2:
			if b.HasError {
				fmt.Fprintf(w, "%s\t%s, err = %s\n", b.Indent, b.Name, callRHS)
			} else if b.HasCleanup {
				fmt.Fprintf(w, "%s\t%s, cleanup = %s\n", b.Indent, b.Name, callRHS)
			} else {
				return fmt.Errorf("invalid signature(2) %s, supported return type are (<T>, error) and (<T>, func(), %w", b.Provider, ErrUnexpectedReturnType)
			}
		case 3:
			if b.HasError && b.HasCleanup {
				fmt.Fprintf(w, "%s\t%s, cleanup, err = %s\n", b.Indent, b.Name, callRHS)
			} else {
				return fmt.Errorf("invalid signature(3) %s, supported return type are (<T>, func(), error), %w", b.Provider, ErrUnexpectedReturnType)
			}
		default:
			return fmt.Errorf("invalid signature(N) %s, %w", b.Provider, ErrUnexpectedReturnType)
		}

		if b.HasCleanup {
			fmt.Fprintf(w, "%s\tif cleanup != nil {\n", b.Indent)
			fmt.Fprintf(w, "%s\t\tdefer cleanup()\n", b.Indent)
			fmt.Fprintf(w, "%s\t}\n", b.Indent)
		}
		if b.HasError { // TODO: support zero-value
			var returnRHS string
			if len(returns) == 0 {
				returnRHS = "panic(err) // TODO: fix-it"
			} else {
				values := []string{"nil", "nil", "nil"}
				if returns[len(returns)-1].Node.String() == "error" {
					values[len(returns)-1] = "err"
				}
				returnRHS = "return " + strings.Join(values[:len(returns)], ", ")
			}

			fmt.Fprintf(w, "%s\tif err != nil {\n", b.Indent)
			fmt.Fprintf(w, "%s\t\t%s\n", b.Indent, returnRHS)
			fmt.Fprintf(w, "%s\t}\n", b.Indent)
		}
	}
	return nil
}
