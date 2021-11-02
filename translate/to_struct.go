package translate

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/namelib"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	reflectshape "github.com/podhmo/reflect-shape"
)

// TranslateToInterface translates to interface from concrete struct
func (t *Translator) TranslateToStruct(here *tinypkg.Package, ob interface{}, name string) *code.CodeEmitter {
	shape := t.Resolver.Shape(ob)
	if name == "" {
		name = shape.GetName()
	}
	var s *Struct // bound by EmitCode()
	c := &code.Code{
		Name:   name,
		Here:   here,
		Config: t.Config,
		ImportPackages: func(collector *tinypkg.ImportCollector) error {
			return s.OnWalk(collector.Collect)
		},
		EmitCode: func(w io.Writer, c *code.Code) error {
			resolver := t.Resolver
			here := c.Here
			switch shape := shape.(type) {
			case reflectshape.Struct:
				s = toStruct(here, resolver, name, shape)
				if err := writeStruct(w, here, name, s, resolver); err != nil {
					return fmt.Errorf("write struct: %w", err)
				}
				return nil
			case reflectshape.Function:
				structShape, err := resolve.StructFromShape(resolver, shape)
				if err != nil {
					return fmt.Errorf("transform function to struct: %w", err)
				}
				s = toStruct(here, resolver, name, structShape)
				if err := writeStruct(w, here, name, s, resolver); err != nil {
					return err
				}
				return nil
			default:
				return fmt.Errorf("%s is not struct or pointer of struct", shape)
			}
		},
	}
	return &code.CodeEmitter{Code: c}
}

func writeStruct(
	w io.Writer,
	here *tinypkg.Package,
	name string,
	s *Struct,
	resolver *resolve.Resolver,
) error {
	// struct {
	// ..
	// }
	fmt.Fprintf(w, "type %s struct {\n", name)
	defer fmt.Fprintln(w, "}")
	for _, field := range s.Fields {
		if len(field.Tags) == 0 {
			fmt.Fprintf(w, "\t%s %s\n", field.Name, tinypkg.ToRelativeTypeString(here, field.Type))
		} else {
			fmt.Fprintf(w, "\t%s %s `%s`\n", field.Name, tinypkg.ToRelativeTypeString(here, field.Type), strings.Join(field.Tags, " "))
		}
	}
	return nil
}

type Struct struct {
	Name    string
	Package *tinypkg.Package
	Fields  []StructField
}

func (s *Struct) OnWalk(use func(*tinypkg.Symbol) error) error {
	for _, f := range s.Fields {
		if err := f.Type.OnWalk(use); err != nil {
			return fmt.Errorf("walk field %s: %w", f.Name, err)
		}
	}
	return nil
}

type StructField struct {
	Name string
	Type tinypkg.Node
	Tags []string
}

func toStruct(here *tinypkg.Package, resolver *resolve.Resolver, name string, s reflectshape.Struct) *Struct {
	fields := make([]StructField, s.Fields.Len())
	for i, fieldname := range s.Fields.Keys {
		f := s.Fields.Values[i]
		var tags []string
		if s.Tags[i] != "" {
			tags = append(tags, string(s.Tags[i]))
		}
		if _, ok := s.Tags[i].Lookup("json"); !ok {
			tags = append(tags, fmt.Sprintf("json:"+strconv.Quote(fieldname)))
		}
		fields[i] = StructField{
			Name: namelib.ToExported(fieldname),
			Type: resolver.Symbol(here, f),
			Tags: tags,
		}
	}
	return &Struct{
		Name:    name,
		Package: here,
		Fields:  fields,
	}
}
