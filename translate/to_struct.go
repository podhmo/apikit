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

type ToStructConfig struct {
	name    string
	tagsMap map[string]string
}
type ToStructOption func(*ToStructConfig)

type ToStructNamespace struct{}

func (t *Translator) OptionToStruct() *ToStructNamespace {
	return nil // for namespace, so, actual value is not needed.
}

func (ns *ToStructNamespace) WithName(name string) ToStructOption {
	return func(c *ToStructConfig) { c.name = name }
}
func (ns *ToStructNamespace) WithTag(pairs ...string) ToStructOption {
	return func(c *ToStructConfig) {
		for i := 0; i+1 < len(pairs); i += 2 {
			c.tagsMap[pairs[i]] = pairs[i+1]
		}
	}
}

// TranslateToStruct translates to struct from function or concrete struct
func (t *Translator) TranslateToStruct(
	here *tinypkg.Package,
	ob interface{},
	options ...ToStructOption,
) *code.CodeEmitter {
	shape := t.Resolver.Shape(ob)
	cfg := &ToStructConfig{tagsMap: map[string]string{}}
	for _, opt := range options {
		opt(cfg)
	}
	if cfg.name == "" {
		cfg.name = shape.GetName()
	}

	name := cfg.name
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
				s = toStruct(here, resolver, shape, *cfg)
				if err := writeStruct(w, here, name, s, resolver); err != nil {
					return fmt.Errorf("write struct: %w", err)
				}
				return nil
			case reflectshape.Function:
				structShape, _, err := resolve.StructFromShape(resolver, shape)
				if err != nil {
					return fmt.Errorf("transform function to struct: %w", err)
				}
				s = toStruct(here, resolver, structShape, *cfg)
				if err := writeStruct(w, here, name, s, resolver); err != nil {
					return fmt.Errorf("write struct: %w", err)
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
		parts := make([]string, 0, 3)
		if field.Embedded {
			parts = append(parts, tinypkg.ToRelativeTypeString(here, field.Type))
		} else {
			parts = append(parts, field.Name)
			parts = append(parts, tinypkg.ToRelativeTypeString(here, field.Type))
			if len(field.Tags) > 0 {
				parts = append(parts, fmt.Sprintf("`%s`", strings.Join(field.Tags, " ")))
			}
		}
		fmt.Fprintf(w, "\t%s\n", strings.Join(parts, " "))
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
	Name     string
	Type     tinypkg.Node
	Tags     []string
	Embedded bool
}

func toStruct(
	here *tinypkg.Package,
	resolver *resolve.Resolver,
	s reflectshape.Struct,
	cfg ToStructConfig,
) *Struct {
	name := cfg.name
	tagsMap := cfg.tagsMap

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
		if additional, ok := tagsMap[fieldname]; ok {
			tags = append(tags, additional)
		}
		fields[i] = StructField{
			Name:     namelib.ToExported(fieldname),
			Type:     resolver.Symbol(here, f),
			Tags:     tags,
			Embedded: s.Metadata[i].Anonymous,
		}
	}
	return &Struct{
		Name:    name,
		Package: here,
		Fields:  fields,
	}
}
