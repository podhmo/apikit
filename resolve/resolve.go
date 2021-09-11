package resolve

import (
	"reflect"

	"github.com/podhmo/apikit/tinypkg"
	"github.com/podhmo/reflect-shape"
	"github.com/podhmo/reflect-shape/arglist"
)

type Resolver struct {
	extractor *reflectshape.Extractor
	universe  *tinypkg.Universe
}

func NewResolver() *Resolver {
	e := &reflectshape.Extractor{
		Seen:           map[reflect.Type]reflectshape.Shape{},
		ArglistLookup:  arglist.NewLookup(),
		RevisitArglist: true,
	}
	return &Resolver{
		extractor: e,
		universe:  tinypkg.NewUniverse(),
	}
}

func (r *Resolver) Resolve(fn interface{}) *Def {
	sfn := r.extractor.Extract(fn).(reflectshape.Function)
	pkg := r.universe.NewPackage(sfn.Package, "")
	args := make([]Item, 0, len(sfn.Params.Keys))

	for i, name := range sfn.Params.Keys {
		s := sfn.Params.Values[i]
		var kind Kind

		if s.GetLv() > 0 {
			kind = KindComponent
		} else {
			switch s := s.(type) {
			case reflectshape.Primitive:
				kind = KindPrimitive
			case reflectshape.Interface:
				if s.GetFullName() == "context.Context" {
					kind = KindIgnored
				} else {
					kind = KindComponent
				}
			case reflectshape.Struct:
				kind = KindData
			case reflectshape.Container: // slice, map
				kind = KindUnsupported
			case reflectshape.Function:
				kind = KindComponent
			default:
				kind = KindUnsupported
			}
		}

		args = append(args, Item{
			Kind:  kind,
			Name:  name,
			Shape: s,
		})
	}
	return &Def{
		Symbol: pkg.NewSymbol(sfn.Name),
		Shape:  sfn,
		Args:   args,
	}
}

type Def struct {
	*tinypkg.Symbol
	Shape reflectshape.Function
	Args  []Item
}

type Item struct {
	Kind  Kind
	Name  string
	Shape reflectshape.Shape
}

type Kind string

const (
	KindComponent   Kind = "component"   // pointer, function, interface
	KindData        Kind = "data"        // struct
	KindIgnored     Kind = "ignoerd"     // context.Context
	KindPrimitive   Kind = "primitve"    // string, int, ...
	KindUnsupported Kind = "unsupported" // slice, map
)
