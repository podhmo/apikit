package enum

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/ext"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

// todo: int enum

type EnumSet struct {
	Name  string
	Enums []*Enum
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
	enums := make([]*Enum, len(members))
	for i, name := range members {
		enums[i] = &Enum{Name: name}
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
	typename := enumSet.Name
	if len(enumSet.Enums) == 0 {
		return fmt.Errorf("need len(enums) > 1, in enum %s", enumSet.Name)
	}

	// fix up
	for _, x := range enumSet.Enums {
		// if value is not set, treated as string-enum
		if x.Value == nil {
			x.Value = x.Name
		}
	}

	rt := reflect.TypeOf(enumSet.Enums[0].Value)
	baseT := resolver.Symbol(here, resolver.Shape(enumSet.Enums[0].Value))

	// <enumset.name>.go
	c := &code.CodeEmitter{Code: config.NewCode(here, enumSet.Name, func(w io.Writer, c *code.Code) error {
		fmt.Fprintf(w, "type %s %s\n", enumSet.Name, baseT)
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "const (")
		for _, x := range enumSet.Enums {

			if x.Description != "" {
				lines := strings.Split(x.Description, "\n")
				lines[0] = fmt.Sprintf("%[1]s%[2]s: %s", typename, x.Name, lines[0])
				for _, line := range lines {
					fmt.Fprintf(w, "\t// %s\n", line)
				}
			}

			switch rt.Kind() {
			case reflect.Int, reflect.Int32, reflect.Int16, reflect.Int64, reflect.Int8:
				fmt.Fprintf(w, "\t%[1]s%[2]s %[1]s = %[3]v\n", typename, x.Name, x.Value)
			case reflect.Uint, reflect.Uint32, reflect.Uint16, reflect.Uint64, reflect.Uint8:
				fmt.Fprintf(w, "\t%[1]s%[2]s %[1]s = %[3]v\n", typename, x.Name, x.Value)
			case reflect.String:
				fmt.Fprintf(w, "\t%[1]s%[2]s %[1]s = %[3]q\n", typename, x.Name, x.Value)
			default:
				return fmt.Errorf("unexpected type %v, kind=%v", rt, rt.Kind())
			}
		}
		fmt.Fprintln(w, ")")

		// todo: zero value
		return nil
	})}
	emitter.Register(here, c.Name, c)
	return nil
}
