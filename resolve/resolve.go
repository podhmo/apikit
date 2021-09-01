package resolve

import (
	"reflect"

	"github.com/podhmo/apikit/tinypkg"
	"github.com/podhmo/reflect-openapi/pkg/arglist"
	"github.com/podhmo/reflect-openapi/pkg/shape"
)

type Resolver struct {
	extractor *shape.Extractor
	universe  *tinypkg.Universe
}

func NewResolver() *Resolver {
	e := &shape.Extractor{Seen: map[reflect.Type]shape.Shape{}, ArglistLookup: arglist.NewLookup()}
	return &Resolver{
		extractor: e,
		universe:  tinypkg.NewUniverse(),
	}
}

func (r *Resolver) Resolve(fn interface{}) *Def {
	sfn := r.extractor.Extract(fn).(shape.Function)
	pkg := r.universe.NewPackage(sfn.Package, "")
	args := make([]Item, 0, len(sfn.Params.Keys))

	for i, name := range sfn.Params.Keys {
		s := sfn.Params.Values[i]
		var kind Kind

		if s.GetLv() > 0 {
			kind = KindComponent
		} else {
			switch s := s.(type) {
			case shape.Primitive:
				kind = KindPrimitive
			case shape.Interface:
				if s.GetFullName() == "context.Context" {
					kind = KindIgnored
				} else {
					kind = KindComponent
				}
			case shape.Struct:
				kind = KindData
			case shape.Container: // slice, map
				kind = KindUnsupported
			case shape.Function:
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
	Shape shape.Function
	Args  []Item
}

type Item struct {
	Kind  Kind
	Name  string
	Shape shape.Shape
	sym   *tinypkg.Symbol
}

func (i *Item) Symbol() *tinypkg.Symbol {
	if i.sym != nil {
		return i.sym
	}
	i.sym = i.extractSymbol()
	return i.sym
}
func (i *Item) extractSymbol() *tinypkg.Symbol {
	return nil // TODO: implement
}

type Kind string

const (
	KindComponent   Kind = "component"   // pointer, function, interface
	KindData        Kind = "data"        // struct
	KindIgnored     Kind = "ignoerd"     // context.Context
	KindPrimitive   Kind = "primitve"    // string, int, ...
	KindUnsupported Kind = "unsupported" // slice, map
)
