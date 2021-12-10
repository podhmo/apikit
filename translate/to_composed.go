package translate

import (
	"fmt"
	"io"
	"strings"

	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	reflectshape "github.com/podhmo/reflect-shape"
)

// wire-like factory composision
func writeComposed(
	w io.Writer,
	here *tinypkg.Package, resolver *resolve.Resolver,
	name string, providers []*resolve.Def,
) error {
	// TODO: external
	type Var struct {
		*tinypkg.Var
		K reflectshape.Identity
	}

	vars := map[reflectshape.Identity]Var{}

	for i, p := range providers {
		retItem := p.Returns[0].Shape
		k := retItem.GetIdentity()
		if _, ok := vars[k]; ok {
			continue
		}
		name := fmt.Sprintf("v%d", i)
		node := resolver.Symbol(here, retItem)
		vars[k] = Var{K: k, Var: &tinypkg.Var{Name: name, Node: node}}
	}

	var args []*tinypkg.Var
	for _, p := range providers {
		for _, x := range p.Args {
			k := x.Shape.GetIdentity()
			v, ok := vars[k]
			if !ok {
				sym := resolver.Symbol(here, x.Shape)
				args = append(args, &tinypkg.Var{Name: x.Name, Node: sym})
				v = Var{K: k, Var: &tinypkg.Var{Name: x.Name, Node: sym}}
				vars[k] = v
			}
			v.Name = x.Name
		}
	}

	var bindings tinypkg.BindingList
	for _, p := range providers {
		retK := p.Returns[0].Shape.GetIdentity()
		v := vars[retK]

		sym := resolver.Symbol(here, p.Shape)
		factory, ok := sym.(*tinypkg.Func)
		if !ok {
			// func() <component>
			factory = here.NewFunc("", nil, []*tinypkg.Var{{Node: sym}})
		}

		b, err := tinypkg.NewBinding(v.Name, factory)
		if err != nil {
			return fmt.Errorf("on provider %s, new binding: %w", p.Symbol, err)
		}

		// handling pointer (ref, deref)
		args := make([]string, len(b.Provider.Args))
		for i, x := range b.Provider.Args {
			v := vars[p.Args[i].Shape.GetIdentity()]
			name := v.Name
			varLv := tinypkg.GetLevel(v.Node)
			argLV := tinypkg.GetLevel(x.Node)
			// fmt.Println("@@", "arg", i, x, tinypkg.GetLevel(x.Node))
			// fmt.Println("@#", "var", v.Name, v.Node, tinypkg.GetLevel(v.Node))
			diff := varLv - argLV
			if diff < 0 {
				name = strings.Repeat("&", diff) + name
			} else if diff > 0 {
				name = strings.Repeat("*", diff) + name
			}
			args[i] = name
		}
		b.ArgsAliases = args
		bindings = append(bindings, b)
	}

	// bindings.Dump(args...)

	sorted, err := bindings.TopologicalSorted(args...)
	if err != nil {
		return fmt.Errorf("topological sort: %w", err)
	}

	returns := sorted[len(sorted)-1].Provider.Returns
	return tinypkg.WriteFunc(w, here, name, &tinypkg.Func{Args: args, Returns: returns},
		func() error {
			indent := "\t"
			for _, b := range sorted[:len(sorted)-1] {
				if err := b.WriteWithCleanupAndError(w, here, indent, returns); err != nil {
					return err
				}
			}
			b := sorted[len(sorted)-1]
			fmt.Fprintf(w, "%sreturn %s\n", indent, b.CallString(here))
			return nil
		})
}
