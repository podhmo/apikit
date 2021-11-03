package enum

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/ext"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/namelib"
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

type Options struct {
	EnumSet EnumSet
}

func (o Options) IncludeMe(pc *ext.PluginContext, here *tinypkg.Package) error {
	return IncludeMe(
		pc.Config, pc.Resolver, pc.Emitter,
		here,
		Fixup(o.EnumSet),
	)
}

func StringEnums(name string, first string, members ...string) EnumSet {
	members = append([]string{first}, members...)
	enums := make([]Enum, len(members))
	for i, name := range members {
		enums[i] = Enum{Name: name}
	}

	return EnumSet{
		Name:  name,
		Enums: enums,
	}
}

func Fixup(enumSet EnumSet) EnumSet {
	normalized := make([]Enum, len(enumSet.Enums))
	for i, x := range enumSet.Enums {
		x := x // copied
		// if value is not set, treated as string-enum
		if x.Value == nil {
			x.Value = x.Name
		}
		x.Name = namelib.ToExported(x.Name)
		normalized[i] = x
	}
	return EnumSet{Name: enumSet.Name, Enums: normalized}
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

	rt := reflect.TypeOf(enumSet.Enums[0].Value)
	baseT := resolver.Symbol(here, resolver.Shape(enumSet.Enums[0].Value))

	here.NewSymbol(enumSet.Name)

	// <enumset.name>.go
	c := &code.CodeEmitter{Code: config.NewCode(here, enumSet.Name, func(w io.Writer, c *code.Code) error {
		c.Import(resolver.NewPackage("fmt", ""))

		// todo: zero value
		// todo: list members

		// type <type> <base>
		fmt.Fprintf(w, "type %s %s\n", enumSet.Name, baseT)
		fmt.Fprintln(w, "")

		// const
		{
			// const (
			//   <name0> <type> = <value0>
			//   ...
			// )
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
		}

		// error value
		fmt.Fprintf(w, "var ErrNo%[1]s = fmt.Errorf(\"no %[1]s\")\n", typename)
		fmt.Fprintln(w, "")

		// validate
		{
			// func(v <type>) Validate() error {
			//   switch v {
			// 	  case <members ...>:
			//      return nil
			// 	  default:
			//      return fmt.Errorf("unexpected value %v: %w", v, ErrNo<name>)
			//   }
			// }

			fmt.Fprintf(w, "func (v %s) Validate() error {\n", typename)
			members := make([]string, len(enumSet.Enums))
			for i, x := range enumSet.Enums {
				members[i] = typename + x.Name
			}

			fmt.Fprintf(w, "\tswitch v {\n")
			fmt.Fprintf(w, "\tcase %s:\n", strings.Join(members, ", "))
			fmt.Fprintln(w, "\t\treturn nil")
			fmt.Fprintln(w, "\tdefault:")
			fmt.Fprintf(w, "\t\treturn fmt.Errorf(\"unexpected value %%v: %%w\", ErrNo%s)\n", typename)
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "}")
		}

		fmt.Fprintln(w, "")
		// must
		{
			// func Must<type>(v <base>) <type> {
			// }
			fmt.Fprintf(w, "func Must%[1]s(v %[2]s) %[1]s {\n", typename, baseT)
			fmt.Fprintf(w, "\tretval := %s(v)\n", typename)
			fmt.Fprintln(w, "\tif err := retval.Validate(); err != nil {")
			fmt.Fprintln(w, "\t\tpanic(err)")
			fmt.Fprintln(w, "\t}")
			fmt.Fprintln(w, "\treturn retval")
			fmt.Fprintln(w, "}")
		}

		fmt.Fprintln(w, "")

		// list
		{
			members := make([]string, len(enumSet.Enums))
			for i, x := range enumSet.Enums {
				members[i] = typename + x.Name
			}

			// func List<type>() []<type> {
			// return <members...>
			// }
			fmt.Fprintf(w, "func List%[1]s() []%[1]s {\n", typename)
			fmt.Fprintf(w, "\treturn []%s{%s}\n", typename, strings.Join(members, ", "))
			fmt.Fprintln(w, "}")
		}
		return nil
	})}
	emitter.Register(here, c.Name, c)
	return nil
}
