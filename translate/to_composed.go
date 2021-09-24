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
	var args []*tinypkg.Var
	var returns []*tinypkg.Var
	symbols := map[reflectshape.Identity]*tinypkg.Symbol{}

	for i, p := range providers {
		k := p.Returns[0].Shape.GetIdentity()
		symbols[k] = here.NewSymbol(fmt.Sprintf("v%d", i))
	}
	for _, p := range providers {
		for _, x := range p.Args {
			k := x.Shape.GetIdentity()
			_, ok := symbols[k]
			if !ok {
				args = append(args, &tinypkg.Var{Name: x.Name, Node: resolver.Symbol(here, x.Shape)})
				symbols[k] = here.NewSymbol(x.Name)
			}
		}
	}

	{
		sym := resolver.Symbol(here, providers[len(providers)-1].Shape)
		f, ok := sym.(*tinypkg.Func)
		if !ok {
			return fmt.Errorf("unexpected return type providers[%d] (%s)", len(providers)-1, sym)
		}
		returns = f.Returns
	}
	return tinypkg.WriteFunc(w, here, name, &tinypkg.Func{Args: args, Returns: returns},
		func() error {
			for i, p := range providers {
				if i == len(providers)-1 {
					args := make([]string, 0, len(p.Args))
					for _, x := range p.Args {
						k := x.Shape.GetIdentity()
						v := symbols[k]
						args = append(args, v.Name)
					}
					rhs := fmt.Sprintf("%s(%s)", tinypkg.ToRelativeTypeString(here, p.Symbol), strings.Join(args, ", "))
					fmt.Fprintf(w, "\treturn %s\n", rhs)
					break
				}

				retK := p.Returns[0].Shape.GetIdentity()
				name := symbols[retK].Name

				sym := resolver.Symbol(here, p.Shape)
				factory, ok := sym.(*tinypkg.Func)
				if !ok {
					// func() <component>
					factory = here.NewFunc("", nil, []*tinypkg.Var{{Node: sym}})
				}
				binding, err := tinypkg.NewBinding(name, factory)
				if err != nil {
					return fmt.Errorf("on provider %s, new binding: %w", p.Symbol, err)
				}
				indent := "\t"
				if err := binding.WriteWithCleanupAndError(w, here, indent, returns); err != nil {
					return fmt.Errorf("on provider %s, write binding: %w", p.Symbol, err)
				}
			}
			return nil
		})
}
