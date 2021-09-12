package resolve

import (
	"fmt"
	"strconv"

	"github.com/podhmo/apikit/tinypkg"
	reflectshape "github.com/podhmo/reflect-shape"
)

func ExtractSymbol(here *tinypkg.Package, s reflectshape.Shape) tinypkg.Node {
	lv := s.GetLv()
	if lv == 0 {
		return extractSymbol(here, s)
	}
	return &tinypkg.Pointer{Lv: lv, V: extractSymbol(here, s)}
}

func extractSymbol(here *tinypkg.Package, s reflectshape.Shape) tinypkg.Node {
	switch s := s.(type) {
	case reflectshape.Primitive, reflectshape.Interface, reflectshape.Struct:
		if s.GetPackage() == "" { // e.g. string, bool, error
			return tinypkg.NewSymbol(s.GetName())
		}
		pkg := tinypkg.NewPackage(s.GetPackage(), "")
		sym := pkg.NewSymbol(s.GetName())
		return here.Import(pkg).Lookup(sym)
	case reflectshape.Container: // slice, map
		name := s.GetName()
		switch name {
		case "map":
			k := ExtractSymbol(here, s.Args[0])
			v := ExtractSymbol(here, s.Args[1])
			return &tinypkg.Map{K: k, V: v}
		case "slice":
			v := ExtractSymbol(here, s.Args[0])
			return &tinypkg.Slice{V: v}
		default:
			panic(fmt.Sprintf("unsupported container shape %+v", s))
		}
	case reflectshape.Function:
		args := make([]*tinypkg.Var, 0, s.Params.Len())
		{
			// TODO: this is shape package's feature (feature request)
			hasName := false
			for i, name := range s.Params.Keys {
				if "arg"+strconv.Itoa(i) != name {
					hasName = true
					break
				}
			}
			for i, name := range s.Params.Keys {
				arg := s.Params.Values[i]
				if !hasName {
					name = ""
				}
				args = append(args, &tinypkg.Var{Name: name, Node: ExtractSymbol(here, arg)})
			}
		}
		returns := make([]*tinypkg.Var, 0, s.Returns.Len())
		{
			hasName := false
			for i, name := range s.Returns.Keys {
				if "ret"+strconv.Itoa(i) != name {
					hasName = true
					break
				}
			}
			for i, name := range s.Returns.Keys {
				arg := s.Returns.Values[i]
				if !hasName {
					name = ""
				}
				returns = append(returns, &tinypkg.Var{Name: name, Node: ExtractSymbol(here, arg)})
			}
		}
		return &tinypkg.Func{
			Name:    s.GetName(),
			Args:    args,
			Returns: returns,
		}
	default:
		panic(fmt.Sprintf("unsupported shape %+v", s))
	}
}
