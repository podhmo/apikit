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
	i := 0
	for _, p := range providers {
		k := p.Returns[0].Shape.GetIdentity()
		symbols[k] = here.NewSymbol(fmt.Sprintf("v%d", i))
		i++
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
	return tinypkg.WriteFunc(w, name, &tinypkg.Func{Args: args, Returns: returns},
		func() error {
			var retK reflectshape.Identity
			for _, p := range providers {
				retK = p.Returns[0].Shape.GetIdentity()
				lhs := symbols[retK] // TODO: handling, multiple values

				args := make([]string, 0, len(p.Args))
				for _, x := range p.Args {
					k := x.Shape.GetIdentity()
					v := symbols[k]
					args = append(args, v.Name)
				}
				rhs := fmt.Sprintf("%s(%s)", tinypkg.ToRelativeTypeString(here, p.Symbol), strings.Join(args, ", "))
				fmt.Fprintf(w, "\t%s := %s\n", lhs, rhs)
				i++
			}
			fmt.Fprintf(w, "\treturn %s\n", symbols[retK].Name)
			return nil
		})
}
