package enum

import (
	"fmt"
	"io"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/ext"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

// todo: int enum

type EnumSet struct {
	Name  string
	Enums []Enum
}

type Enum struct {
	Name        string
	Value       interface{}
	Description string
}

type Options EnumSet

func (o Options) IncludeMe(pc *ext.PluginContext, here *tinypkg.Package) error {
	return IncludeMe(
		pc.Config, pc.Resolver, pc.Emitter,
		here,
		EnumSet(o),
	)
}

func StringEnums(name string, first string, members ...string) Options {
	members = append([]string{first}, members...)
	enums := make([]Enum, len(members))
	for i, name := range members {
		enums[i] = Enum{Name: name}
	}

	return Options{
		Name:  name,
		Enums: enums,
	}
}

func IncludeMe(
	config *code.Config, resolver *resolve.Resolver, emitter *emitgo.Emitter,
	here *tinypkg.Package,
	enumSet EnumSet,
) error {
	if len(enumSet.Enums) == 0 {
		return fmt.Errorf("need len(enums) > 1, in enum %s", enumSet.Name)
	}
	val := enumSet.Enums[0].Value
	if val == nil { // if value is not set, treated as string-enum
		val = enumSet.Enums[0].Name
	}
	baseT := resolver.Symbol(here, resolver.Shape(val))

	// <enumset.name>.go
	c := &code.CodeEmitter{Code: config.NewCode(here, enumSet.Name, func(w io.Writer, c *code.Code) error {
		fmt.Fprintf(w, "type %s %s", enumSet.Name, baseT)
		return nil
	})}
	emitter.Register(here, c.Name, c)
	return nil
}
