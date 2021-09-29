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
	Name string

	Provider      *Func
	ProviderAlias string

	ZeroReturnsDefault string

	HasError   bool
	HasCleanup bool
}

var ErrUnexpectedReturnType = fmt.Errorf("unexpected-return-type")
var ErrUnexpectedExternalReturnType = fmt.Errorf("unexpected-external-return-type")

func NewBinding(name string, provider *Func) (*Binding, error) {
	b := &Binding{
		Name:               name,
		Provider:           provider,
		ZeroReturnsDefault: "panic(err) // TODO: fix-it",
	}
	switch len(provider.Returns) {
	case 1:
		// noop
	case 2:
		if provider.Returns[1].Node.String() == "error" {
			b.HasError = true
		} else if _, ok := provider.Returns[1].Node.(*Func); ok {
			// TODO: validation
			b.HasCleanup = true
		} else {
			return nil, fmt.Errorf("invalid signature(2) %s, supported return type are (<T>, error) and (<T>, func(), %w", provider, ErrUnexpectedReturnType)
		}
	case 3:
		if provider.Returns[2].Node.String() == "error" {
			b.HasError = true
		}
		if _, ok := provider.Returns[1].Node.(*Func); ok {
			// TODO: validation
			b.HasCleanup = true
		}
		if !(b.HasError && b.HasCleanup) {
			return nil, fmt.Errorf("invalid signature(3) %s, supported return type are (<T>, func(), error), %w", b.Provider, ErrUnexpectedReturnType)
		}
	default:
		return nil, fmt.Errorf("invalid signature(N) %s, %w", provider, ErrUnexpectedReturnType)
	}
	return b, nil
}

// TODO: support non-pointer zero value
// TODO: name-check (when calling provider function)

func (b *Binding) WriteWithCleanupAndError(w io.Writer, here *Package, indent string, returns []*Var) error {
	if 3 < len(returns) {
		return fmt.Errorf("sorry the maximum value of supported number-of-return-value is 3, but %s is passed, %w", returns, ErrUnexpectedExternalReturnType)
	}

	fmt.Fprintf(w, "%svar %s %s\n", indent, b.Name, ToRelativeTypeString(here, b.Provider.Returns[0].Node))
	fmt.Fprintf(w, "%s{\n", indent)
	defer fmt.Fprintf(w, "%s}\n", indent)
	{
		if b.HasCleanup {
			fmt.Fprintf(w, "%s\tvar cleanup func()\n", indent)
		}
		if b.HasError {
			fmt.Fprintf(w, "%s\tvar err error\n", indent)
		}

		var callRHS string
		{
			provider := b.Provider
			args := make([]string, 0, len(provider.Args))
			for _, x := range provider.Args {
				args = append(args, x.Name)
			}
			providerName := provider.Name
			if b.ProviderAlias != "" {
				providerName = b.ProviderAlias
			}
			callRHS = fmt.Sprintf("%s(%s)", providerName, strings.Join(args, ", "))
		}

		switch len(b.Provider.Returns) {
		case 1:
			fmt.Fprintf(w, "%s\t%s = %s\n", indent, b.Name, callRHS)
		case 2:
			if b.HasError {
				fmt.Fprintf(w, "%s\t%s, err = %s\n", indent, b.Name, callRHS)
			} else if b.HasCleanup {
				fmt.Fprintf(w, "%s\t%s, cleanup = %s\n", indent, b.Name, callRHS)
			} else {
				return fmt.Errorf("invalid signature(2) %s, supported return type are (<T>, error) and (<T>, func(), %w", b.Provider, ErrUnexpectedReturnType)
			}
		case 3:
			if b.HasError && b.HasCleanup {
				fmt.Fprintf(w, "%s\t%s, cleanup, err = %s\n", indent, b.Name, callRHS)
			} else {
				return fmt.Errorf("invalid signature(3) %s, supported return type are (<T>, func(), error), %w", b.Provider, ErrUnexpectedReturnType)
			}
		default:
			return fmt.Errorf("invalid signature(N) %s, %w", b.Provider, ErrUnexpectedReturnType)
		}

		if b.HasCleanup {
			fmt.Fprintf(w, "%s\tif cleanup != nil {\n", indent)
			fmt.Fprintf(w, "%s\t\tdefer cleanup()\n", indent)
			fmt.Fprintf(w, "%s\t}\n", indent)
		}
		if b.HasError { // TODO: support zero-value
			var returnRHS string
			if len(returns) == 0 {
				returnRHS = b.ZeroReturnsDefault
				if returnRHS == "" {
					returnRHS = "panic(err) // TODO: fix-it"
				}
			} else {
				values := []string{"nil", "nil", "nil"}
				if returns[len(returns)-1].Node.String() == "error" {
					values[len(returns)-1] = "err"
				}
				returnRHS = "return " + strings.Join(values[:len(returns)], ", ")
			}

			fmt.Fprintf(w, "%s\tif err != nil {\n", indent)
			fmt.Fprintf(w, "%s\t\t%s\n", indent, returnRHS)
			fmt.Fprintf(w, "%s\t}\n", indent)
		}
	}
	return nil
}
