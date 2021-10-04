package resolve

import (
	"fmt"

	"github.com/podhmo/apikit/pkg/tinypkg"
	reflectshape "github.com/podhmo/reflect-shape"
)

func ExtractSymbol(universe *tinypkg.Universe, here *tinypkg.Package, s reflectshape.Shape) tinypkg.Node {
	lv := s.GetLv()
	if lv == 0 {
		return extractSymbol(universe, here, s)
	}
	return &tinypkg.Pointer{Lv: lv, V: extractSymbol(universe, here, s)}
}

func extractSymbol(universe *tinypkg.Universe, here *tinypkg.Package, s reflectshape.Shape) tinypkg.Node {
	switch s := s.(type) {
	case reflectshape.Primitive, reflectshape.Struct:
		if s.GetPackage() == "" { // e.g. string, bool
			return tinypkg.NewSymbol(s.GetName())
		}

		pkg := universe.NewPackage(s.GetPackage(), "")
		sym := pkg.NewSymbol(s.GetName())
		return here.Import(pkg).Lookup(sym)

	case reflectshape.Interface:
		name := s.GetName()
		pkg := universe.NewPackage(s.GetPackage(), "")
		if name != "" {
			if s.GetPackage() == "" { // e.g. error
				return tinypkg.NewSymbol(s.GetName())
			}

			sym := pkg.NewSymbol(s.GetName())
			return here.Import(pkg).Lookup(sym)
		}

		methods := make([]*tinypkg.Func, len(s.Methods.Keys))
		for i, methodName := range s.Methods.Keys {
			m := s.Methods.Values[i]
			sym := ExtractSymbol(universe, here, m)
			fn, ok := sym.(*tinypkg.Func)
			if !ok {
				panic(fmt.Sprintf("unexpected method members, %s: %T", methodName, sym))
			}
			fn.Name = methodName
			methods[i] = fn
		}
		return &tinypkg.Interface{Name: name, Package: pkg, Methods: methods}

	case reflectshape.Container: // slice, map
		name := s.GetName()
		switch name {
		case "map":
			k := ExtractSymbol(universe, here, s.Args[0])
			v := ExtractSymbol(universe, here, s.Args[1])
			return &tinypkg.Map{K: k, V: v}
		case "slice":
			v := ExtractSymbol(universe, here, s.Args[0])
			return &tinypkg.Slice{V: v}
		default:
			panic(fmt.Sprintf("unsupported container shape %+v", s))
		}
	case reflectshape.Function:
		args := make([]*tinypkg.Var, 0, s.Params.Len())
		{
			for i, name := range s.Params.Keys {
				arg := s.Params.Values[i]
				args = append(args, &tinypkg.Var{Name: name, Node: ExtractSymbol(universe, here, arg)})
			}
		}
		returns := make([]*tinypkg.Var, 0, s.Returns.Len())
		{
			for i, name := range s.Returns.Keys {
				arg := s.Returns.Values[i]
				returns = append(returns, &tinypkg.Var{Name: name, Node: ExtractSymbol(universe, here, arg)})
			}
		}
		pkg := universe.NewPackage(s.GetPackage(), "")
		return &tinypkg.Func{
			Name:    s.GetName(),
			Package: pkg,
			Args:    args,
			Returns: returns,
		}
	default:
		panic(fmt.Sprintf("unsupported shape %+v", s))
	}
}
