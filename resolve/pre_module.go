package resolve

import (
	"fmt"

	"github.com/podhmo/apikit/pkg/tinypkg"
	reflectshape "github.com/podhmo/reflect-shape"
)

var ErrNotFound = fmt.Errorf("not found")
var ErrInvalidType = fmt.Errorf("invalid type")
var ErrInvalidArgs = fmt.Errorf("invalid args")

type GetSymbolFunc func(*tinypkg.Package, interface{}) tinypkg.Node
type PreModule struct {
	Name string

	Shape *reflectshape.Struct

	Args  []reflectshape.Shape
	Funcs []reflectshape.Function

	resolver *Resolver
}

func NewPreModule(resolver *Resolver, ob reflectshape.Struct) (*PreModule, error) {
	var args []reflectshape.Shape
	var funcs []reflectshape.Function
	for i, f := range ob.Fields.Values {
		// TODO: use tag?
		if f, ok := f.(reflectshape.Function); ok {
			name := ob.Fields.Keys[i]
			f.Name = name
			funcs = append(funcs, f)
			continue
		}
		args = append(args, f)
	}

	return &PreModule{
		Name:     ob.Name,
		Shape:    &ob,
		Args:     args,
		Funcs:    funcs,
		resolver: resolver,
	}, nil
}

func (m *PreModule) String() string {
	var prefix string
	if m.Name != "" {
		prefix = " " + m.Name + ""
	}
	return fmt.Sprintf("PreModule%s[%s]", prefix, m.Args)
}

func (f *PreModule) NewModule(here *tinypkg.Package, args ...tinypkg.Node) (*Module, error) {
	if len(f.Args) != len(args) {
		return nil, fmt.Errorf("the length of args is invalid, expected args is %s, but got %s: %w", f.Args, args, ErrInvalidArgs)
	}

	replaceMap := make(map[*tinypkg.Symbol]tinypkg.Node, len(f.Args))
	for i, arg := range f.Args {
		sym := f.resolver.NewPackage(arg.GetPackage(), "").NewSymbol(arg.GetName())
		replaceMap[sym] = args[i]
	}
	return &Module{
		origin:     f,
		Here:       here,
		Args:       args,
		replaceMap: replaceMap,
		funcs:      make(map[string]*tinypkg.Func, f.Shape.Fields.Len()),
	}, nil
}

type Module struct {
	origin     *PreModule
	replaceMap map[*tinypkg.Symbol]tinypkg.Node

	Args  []tinypkg.Node
	Here  *tinypkg.Package
	funcs map[string]*tinypkg.Func
}

func (m *Module) String() string {
	funcNames := make([]string, 0, len(m.origin.Funcs))
	for _, f := range m.origin.Funcs {
		funcNames = append(funcNames, f.Name)
	}
	return fmt.Sprintf("Module[here=%s, args=%s, funcs=%s]", m.Here.Path, m.Args, funcNames)
}

func (m *Module) Shape() *reflectshape.Struct {
	return m.origin.Shape
}

func (m *Module) Symbol(here *tinypkg.Package, name string) (*tinypkg.ImportedSymbol, error) {
	_, ok := m.origin.Shape.Fields.Get(name)
	if !ok {
		return nil, fmt.Errorf("the function %s is %w", name, ErrNotFound)
	}
	sym := m.Here.NewSymbol(name)
	return here.Import(m.Here).Lookup(sym), nil
}

func (m *Module) Type(name string) (*tinypkg.Func, error) {
	if f, ok := m.funcs[name]; ok {
		return f, nil
	}

	here := m.Here
	shape, ok := m.origin.Shape.Fields.Get(name)
	if !ok {
		return nil, fmt.Errorf("the function %s is %w", name, ErrNotFound)
	}

	sym := m.origin.resolver.Symbol(here, shape)
	fn, ok := sym.(*tinypkg.Func)
	if !ok {
		return nil, fmt.Errorf("the function %s is something wrong (shape=%s): %w", name, shape, ErrNotFound)
	}

	if len(m.Args) > 0 {
		fn = tinypkg.Replace(fn, func(sym *tinypkg.Symbol) tinypkg.Node {
			if replaced, ok := m.replaceMap[sym]; ok {
				return replaced
			}
			return sym
		}).(*tinypkg.Func)
	}
	fn = here.NewFunc(name, fn.Args, fn.Returns)
	m.funcs[name] = fn
	return fn, nil
}
