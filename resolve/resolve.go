package resolve

import (
	"reflect"

	"github.com/podhmo/apikit/tinypkg"
	reflectshape "github.com/podhmo/reflect-shape"
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

func detectKind(s reflectshape.Shape) Kind {
	if s.GetLv() > 0 {
		// TODO: if the pointer of primitive passed, treated as optional value? (not yet)
		return KindComponent
	}

	switch s := s.(type) {
	case reflectshape.Primitive:
		return KindPrimitive
	case reflectshape.Interface:
		if s.GetFullName() == "context.Context" {
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

func (r *Resolver) Resolve(fn interface{}) *Def {
	sfn := r.extractor.Extract(fn).(reflectshape.Function)
	pkg := r.universe.NewPackage(sfn.Package, "")
	args := make([]Item, 0, len(sfn.Params.Keys))
	returns := make([]Item, 0, len(sfn.Returns.Keys))

	for i, name := range sfn.Params.Keys {
		s := sfn.Params.Values[i]
		kind := detectKind(s)
		args = append(args, Item{
			Kind:  kind,
			Name:  name,
			Shape: s,
		})
	}
	for i, name := range sfn.Returns.Keys {
		s := sfn.Returns.Values[i]
		kind := detectKind(s)
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
	KindComponent   Kind = "component"   // pointer, function, interface
	KindData        Kind = "data"        // struct
	KindIgnored     Kind = "ignoerd"     // context.Context
	KindPrimitive   Kind = "primitve"    // string, int, ...
	KindUnsupported Kind = "unsupported" // slice, map
)
