package translate

import (
	"fmt"
	"io"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	reflectshape "github.com/podhmo/reflect-shape"
)

// TranslateToInterface translates to interface from concrete struct
func (t *Translator) TranslateToInterface(here *tinypkg.Package, ob interface{}, name string) *code.Code {
	shape := t.Resolver.Shape(ob)
	if name == "" {
		name = shape.GetName()
	}
	return &code.Code{
		Name:   name,
		Here:   here,
		Config: t.Config,
		// ImportPackages: func() ([]*tinypkg.ImportedPackage, error) {
		// 	return nil, nil // TODO: implement
		// },
		EmitCode: func(w io.Writer) error {
			return writeInterface(w, here, t.Resolver, shape, name)
		},
	}
}

func writeInterface(w io.Writer, here *tinypkg.Package, resolver *resolve.Resolver, shape reflectshape.Shape, name string) error {
	s, ok := shape.(reflectshape.Struct)
	if !ok {
		return fmt.Errorf("%s is not struct or pointer of struct", shape)
	}

	var methods []*tinypkg.Func
	methodMap := s.Methods()
	for _, m := range methodMap.Values {
		m.Params.Keys = make([]string, len(m.Params.Keys)-1)
		m.Params.Values = m.Params.Values[1:]

		sym := resolver.Symbol(here, m) // todo: (method support @reflect-shape )
		f, ok := sym.(*tinypkg.Func)
		if !ok {
			return fmt.Errorf("%s's %s is not function", shape, sym)
		}
		methods = append(methods, f)
	}
	iface := here.NewInterface(name, methods)
	return tinypkg.WriteInterface(w, here, name, iface)
}
