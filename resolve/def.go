package resolve

import (
	"github.com/podhmo/apikit/pkg/tinypkg"
	reflectshape "github.com/podhmo/reflect-shape"
)

func ExtractDef(universe *tinypkg.Universe, extractor *reflectshape.Extractor, fn interface{}, ignoreMap map[string]bool) *Def {
	sfn := extractor.Extract(fn).(reflectshape.Function)
	pkg := universe.NewPackage(sfn.Package, "")
	args := make([]Item, 0, len(sfn.Params.Keys))
	returns := make([]Item, 0, len(sfn.Returns.Keys))

	for i, name := range sfn.Params.Keys {
		s := sfn.Params.Values[i]
		kind := DetectKind(s, ignoreMap)
		args = append(args, Item{
			Kind:  kind,
			Name:  name,
			Shape: s,
		})
	}
	for i, name := range sfn.Returns.Keys {
		s := sfn.Returns.Values[i]
		kind := DetectKind(s, ignoreMap)
		returns = append(returns, Item{
			Kind:  kind,
			Name:  name,
			Shape: s,
		})
	}

	return &Def{
		Symbol:  pkg.NewSymbol(sfn.Name),
		Shape:   sfn,
		Args:    args,
		Returns: returns,
	}
}

type Def struct {
	*tinypkg.Symbol
	Shape   reflectshape.Function
	Args    []Item
	Returns []Item
}

type Item struct {
	Kind  Kind
	Name  string
	Shape reflectshape.Shape
}

type Kind string

const (
	KindComponent        Kind = "component"        // pointer, function, interface
	KindData             Kind = "data"             // struct
	KindIgnored          Kind = "ignoerd"          // context.Context
	KindPrimitive        Kind = "primitve"         // string, int, ...
	KindPrimitivePointer Kind = "primitve-pointer" // *string, *int, ...
	KindUnsupported      Kind = "unsupported"      // slice, map
)

func DetectKind(s reflectshape.Shape, ignoreMap map[string]bool) Kind {
	if s.GetLv() > 0 {
		if _, ok := s.(reflectshape.Primitive); ok {
			return KindPrimitivePointer
		}
		// TODO: if the pointer of primitive passed, treated as optional value? (not yet)
		if ignoreMap[s.GetFullName()] {
			return KindIgnored
		} else {
			return KindComponent
		}
	}

	switch s := s.(type) {
	case reflectshape.Primitive:
		return KindPrimitive
	case reflectshape.Interface:
		if ignoreMap[s.GetFullName()] {
			return KindIgnored
		} else {
			return KindComponent
		}
	case reflectshape.Struct:
		return KindData
	case reflectshape.Container: // slice, map
		return KindUnsupported
	case reflectshape.Function:
		return KindComponent
	default:
		return KindUnsupported
	}
}
