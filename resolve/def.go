package resolve

import (
	"fmt"
	"reflect"

	"github.com/podhmo/apikit/pkg/tinypkg"
	reflectshape "github.com/podhmo/reflect-shape"
)

type Def struct {
	*tinypkg.Symbol
	Shape   reflectshape.Function
	Args    []Item
	Returns []Item

	HasError   bool
	HasCleanup bool
}

type Item struct {
	Kind  Kind
	Name  string
	Shape reflectshape.Shape
}

func ExtractDef(universe *tinypkg.Universe, extractor *reflectshape.Extractor, fn interface{}) *Def {
	sfn := extractor.Extract(fn).(reflectshape.Function)
	pkg := universe.NewPackage(sfn.Package, "")
	args := make([]Item, 0, len(sfn.Params.Keys))
	returns := make([]Item, 0, len(sfn.Returns.Keys))

	for i, name := range sfn.Params.Keys {
		s := sfn.Params.Values[i]
		kind := DetectKind(s)
		args = append(args, Item{
			Kind:  kind,
			Name:  name,
			Shape: s,
		})
	}
	for i, name := range sfn.Returns.Keys {
		s := sfn.Returns.Values[i]
		kind := DetectKind(s)
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

var ErrUnexpectedReturnType = fmt.Errorf("unexpected-return-type")
var reflectErrorTypeValue = reflect.TypeOf(func() error { return nil }).Out(0)

func bindReturnsInfo(def *Def) error {
	returns := def.Returns
	switch len(returns) {
	case 0:
		return nil
	case 1:
		lastValueType := returns[0].Shape.GetReflectType()
		if lastValueType == reflectErrorTypeValue { // error
			def.HasError = true
			return nil
		} else if lastValueType.Kind() == reflect.Func { // (<T>, func())
			def.HasCleanup = true
			// TODO: validation
			return nil
		}
		return nil
	case 2:
		lastValueType := returns[1].Shape.GetReflectType()
		if lastValueType == reflectErrorTypeValue { // (<T>, error)
			def.HasError = true
			return nil
		} else if lastValueType.Kind() == reflect.Func { // (<T>, func())
			def.HasCleanup = true
			// TODO: validation
			return nil
		} else {
			return fmt.Errorf("returns[2] signature invalid (%s): %w", def.Symbol, ErrUnexpectedReturnType)
		}
	case 3: // (<T>, func(), error)
		lastValueType := returns[2].Shape.GetReflectType()
		if lastValueType == reflectErrorTypeValue { // (<T>, error)
			def.HasError = true
		} else {
			return fmt.Errorf("returns[3] signature invalid (%s): %w", def.Symbol, ErrUnexpectedReturnType)
		}

		sndValueTypeKind := returns[1].Shape.GetReflectKind()
		if sndValueTypeKind == reflect.Func {
			def.HasCleanup = true
			// TODO: validation
		} else {
			return fmt.Errorf("returns[3] signature invalid (%s): %w", def.Symbol, ErrUnexpectedReturnType)
		}
		return nil
	default:
		return fmt.Errorf("returns[%d] signature invalid (%s): %w", len(returns), def.Symbol, ErrUnexpectedReturnType)
	}
}

type Kind string

const (
	KindComponent   Kind = "component"   // pointer, function, interface
	KindData        Kind = "data"        // struct
	KindIgnored     Kind = "ignoerd"     // context.Context
	KindPrimitive   Kind = "primitve"    // string, int, ...
	KindUnsupported Kind = "unsupported" // slice, map
)

func DetectKind(s reflectshape.Shape) Kind {
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
