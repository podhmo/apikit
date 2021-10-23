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
	argsAliases   []string // args in call alias

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
		if provider.Returns[0].Node.String() == "error" {
			b.HasError = true
		}
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

// WriteWithCleanaupAndError writes binding-code
// support:
// - func(...) <T>
// - func(...) error
// - func(...) (<T>, error)
// - func(...) (<T>, func())
// - func(...) (<T>, func(), error)
func (b *Binding) WriteWithCleanupAndError(w io.Writer, here *Package, indent string, returns []*Var) error {
	if 3 < len(returns) {
		return fmt.Errorf("sorry the maximum value of supported number-of-return-value is 3, but %s is passed, %w", returns, ErrUnexpectedExternalReturnType)
	}

	// similify: <lhs> := <call-rhs>; return <lhs>

	var callRHS string // <fn>(...)
	{
		provider := b.Provider
		args := b.argsAliases
		if args == nil {
			args = make([]string, 0, len(provider.Args))
			for _, x := range provider.Args {
				args = append(args, x.Name)
			}
		}

		providerName := provider.Name
		if b.ProviderAlias != "" {
			providerName = b.ProviderAlias
		} else if provider.Package != nil && provider.Package != here {
			providerName = ToRelativeTypeString(here, provider.Package.NewSymbol(provider.Name))
		}
		callRHS = fmt.Sprintf("%s(%s)", providerName, strings.Join(args, ", "))
	}

	// TODO: support zero-value
	returnValue := b.ZeroReturnsDefault
	if returnValue == "" {
		returnValue = "panic(err) // TODO: fix-it"
	}

	// special optimization for func(...) error.
	if b.HasError && len(b.Provider.Returns) == 1 {
		// generated code is something like this.
		// if err := <call-rhs>; err != nil {
		//     <return-value>
		// }
		if len(returns) > 0 {
			if returns[len(returns)-1].Node.String() == "error" {
				values := []string{"nil", "nil", "nil"}
				values[len(returns)-1] = "err"
				returnValue = "return " + strings.Join(values[:len(returns)], ", ")
			}
		}
		fmt.Fprintf(w, "%sif err := %s; err != nil {\n", indent, callRHS)
		fmt.Fprintf(w, "%s\t%s\n", indent, returnValue)
		fmt.Fprintf(w, "%s}\n", indent)
		return nil
	}

	// generated code (full-set) is something like this.
	// var <name> <T>
	// {
	//     var cleanup func() // if hasCleanup is true
	//     var err error      // if hasError is true
	//     <name>, cleanup, err = <call-rhs>
	//     if cleanup != nil {
	//	        defer cleanup()
	//     }
	//     if err != nil {
	//	        <return-value>
	//     }
	// }
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
			if len(returns) > 0 {
				values := []string{"nil", "nil", "nil"}
				if returns[len(returns)-1].Node.String() == "error" {
					values[len(returns)-1] = "err"
				}
				returnValue = "return " + strings.Join(values[:len(returns)], ", ")
			}

			fmt.Fprintf(w, "%s\tif err != nil {\n", indent)
			fmt.Fprintf(w, "%s\t\t%s\n", indent, returnValue)
			fmt.Fprintf(w, "%s\t}\n", indent)
		}
	}
	return nil
}

type BindingList []*Binding
type topoState struct {
	sorted []*Binding

	seen  map[string]bool
	deps  map[string][]string
	nodes map[string]*Binding
}

func (bl BindingList) TopologicalSorted() ([]*Binding, error) {
	s := &topoState{
		sorted: make([]*Binding, 0, len(bl)),
		seen:   make(map[string]bool, len(bl)),
		deps:   make(map[string][]string, len(bl)),
		nodes:  make(map[string]*Binding, len(bl)),
	}
	for _, b := range bl {
		var deps []string
		for _, x := range b.Provider.Args {
			deps = append(deps, x.Name) // normalize
		}
		s.nodes[b.Name] = b
		s.deps[b.Name] = deps
	}
	for _, b := range bl {
		if err := bl.topoWalk(s, b, nil); err != nil {
			return s.sorted, err
		}
	}
	return s.sorted, nil
}

func (bl *BindingList) topoWalk(s *topoState, current *Binding, history []*Binding) error {
	history = append(history, current)
	if deps, ok := s.deps[current.Name]; ok {
		for _, name := range deps {
			b, ok := s.nodes[name]
			if !ok {
				return fmt.Errorf("node %q is not found in binding[name=%q, need=%s]", name, current.Name, current.Provider)
			}

			for _, x := range history {
				if x == b {
					history = append(history, x)
					hs := make([]string, len(history))
					for i, y := range history {
						hs[i] = fmt.Sprintf("binding[name=%q, need=%s]", y.Name, y.Provider)
					}
					return fmt.Errorf("circular dependency is detected, history=%v", hs)
				}
			}
			if err := bl.topoWalk(s, s.nodes[name], history); err != nil {
				return err
			}
		}
	}
	if _, ok := s.seen[current.Name]; ok {
		return nil
	}
	s.seen[current.Name] = true
	s.sorted = append(s.sorted, current)
	return nil
}
