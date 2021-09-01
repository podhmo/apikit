package resolve

import (
	"fmt"

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
	case shape.Primitive:
		if s.GetPackage() == "" { // e.g. string, bool
			return tinypkg.NewSymbol(s.GetName())
		}
		pkg := tinypkg.NewPackage(s.GetPackage(), "")
		sym := pkg.NewSymbol(s.GetName())
		return here.Import(pkg).Lookup(sym)
	case shape.Interface, shape.Struct:
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
	// case shape.Function:
	default:
		panic(fmt.Sprintf("unsupported shape %+v", s))
	}
}
