package tinypkg

import (
	"fmt"
	"strings"
)

func ToRelativeTypeString(here *Package, symboler Symboler) string {
	switch x := symboler.(type) {
	case *Var:
		return x.Name + " " + ToRelativeTypeString(here, x.Symboler)
	case *Pointer:
		return strings.Repeat("*", x.Lv) + ToRelativeTypeString(here, x.V)
	case *Array:
		return fmt.Sprintf("[%d]%s", x.N, ToRelativeTypeString(here, x.V))
	case *Slice:
		return fmt.Sprintf("[]%s", ToRelativeTypeString(here, x.V))
	case *Map:
		return fmt.Sprintf("map[%s]%s", ToRelativeTypeString(here, x.K), ToRelativeTypeString(here, x.V))
	case *Func:
		params := make([]string, len(x.Params))
		for i, x := range x.Params {
			params[i] = ToRelativeTypeString(here, x)
		}
		returns := make([]string, len(x.Returns))
		for i, x := range x.Returns {
			returns[i] = ToRelativeTypeString(here, x)
		}

		if len(returns) == 1 {
			return fmt.Sprintf("func(%s) %s", strings.Join(params, ", "), strings.Join(returns, ", "))
		}
		return fmt.Sprintf("func(%s) (%s)", strings.Join(params, ", "), strings.Join(returns, ", "))
	case *Symbol:
		if x.Package.Name == "" {
			return x.Name
		}
		if here == nil {
			return x.Name
		}
		if here == x.Package {
			return x.Name
		}
		return x.Package.Name + "." + x.Name
	case *ImportedSymbol:
		if x.pkg.pkg.Name == "" {
			return x.sym.Name
		}
		if here == x.pkg.pkg {
			return x.sym.Name
		}
		return x.Qualifier() + "." + x.sym.Name
	default:
		panic(fmt.Sprintf("unsupported type %T", symboler))
	}
}
