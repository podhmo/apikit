package resolve

import (
	"fmt"
	"strconv"

	"github.com/podhmo/apikit/tinypkg"
	"github.com/podhmo/reflect-openapi/pkg/shape"
)

func ExtractSymbol(here *tinypkg.Package, s shape.Shape) tinypkg.Symboler {
	lv := s.GetLv()
	if lv == 0 {
		return extractSymbol(here, s)
	}
	return &tinypkg.Pointer{Lv: lv, V: extractSymbol(here, s)}
}

func extractSymbol(here *tinypkg.Package, s shape.Shape) tinypkg.Symboler {
	switch s := s.(type) {
	case shape.Primitive, shape.Interface, shape.Struct:
		if s.GetPackage() == "" { // e.g. string, bool, error
			return tinypkg.NewSymbol(s.GetName())
		}
		pkg := tinypkg.NewPackage(s.GetPackage(), "")
		sym := pkg.NewSymbol(s.GetName())
		return here.Import(pkg).Lookup(sym)
	case shape.Container: // slice, map
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
	case shape.Function:
		params := make([]*tinypkg.Var, 0, s.Params.Len())
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
				params = append(params, &tinypkg.Var{Name: name, Symboler: ExtractSymbol(here, arg)})
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
				returns = append(returns, &tinypkg.Var{Name: name, Symboler: ExtractSymbol(here, arg)})
			}
		}
		return &tinypkg.Func{
			Name:    s.GetName(),
			Params:  params,
			Returns: returns,
		}
	default:
		panic(fmt.Sprintf("unsupported shape %+v", s))
	}
}
