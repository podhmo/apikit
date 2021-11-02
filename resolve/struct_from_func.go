package resolve

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/podhmo/apikit/pkg/namelib"
	"github.com/podhmo/apikit/pkg/tinypkg"
	reflectshape "github.com/podhmo/reflect-shape"
)

func StructFromShape(resolver *Resolver, fn reflectshape.Function) (reflectshape.Shape, error) {
	fields := reflectshape.ShapeMap{}
	tags := make([]reflect.StructTag, 0, fn.Params.Len())
	metadata := make([]reflectshape.FieldMetadata, 0, fn.Params.Len())
	for i, name := range fn.Params.Keys {
		p := fn.Params.Values[i]
		switch kind := resolver.DetectKind(p); kind {
		case KindIgnored, KindUnsupported, KindComponent:
			continue
		case KindData, KindPrimitive, KindPrimitivePointer:
			switch kind {
			case KindData:
				s := p.(reflectshape.Struct)
				fields.Keys = append(fields.Keys, s.Fields.Keys...)
				fields.Values = append(fields.Values, s.Fields.Values...)
				metadata = append(metadata, s.Metadata...)
				tags = append(tags, s.Tags...)
			case KindPrimitive:
				fields.Keys = append(fields.Keys, name)
				fields.Values = append(fields.Values, p)
				metadata = append(metadata, reflectshape.FieldMetadata{
					FieldName: name,
					Required:  true,
				})
			case KindPrimitivePointer:
				fields.Keys = append(fields.Keys, name)
				fields.Values = append(fields.Values, p)
				metadata = append(metadata, reflectshape.FieldMetadata{
					FieldName: name,
					Required:  false,
				})
			}
		default:
			return nil, fmt.Errorf("unsupported kind %v", kind)
		}
	}

	retval := reflectshape.Struct{
		Info: &reflectshape.Info{
			Name:    "", // not ref
			Kind:    reflectshape.Kind(reflect.Struct),
			Package: fn.Info.Package,
		},
		Fields:   fields,
		Tags:     tags,
		Metadata: metadata,
	}
	retval.ResetReflectType(reflect.PtrTo(fn.GetReflectType())) // xxx:
	return retval, nil
}

// TODO: move

type Struct struct {
	Name    string
	Package *tinypkg.Package
	Fields  []StructField
}

type StructField struct {
	Name string
	Type tinypkg.Node
	Tag  string // reflect.StructTag?
}

func toStruct(here *tinypkg.Package, resolver *Resolver, name string, s reflectshape.Struct) *Struct {
	fields := make([]StructField, s.Fields.Len())
	for i, fieldname := range s.Fields.Keys {
		f := s.Fields.Values[i]
		var tag string
		if len(s.Tags) > i {
			tag = string(s.Tags[i])
		}
		if !strings.Contains(tag, "json:") {
			tag = fmt.Sprintf(`json:"%s" %s`, fieldname, tag)
		}
		fields[i] = StructField{
			Name: namelib.ToExported(fieldname),
			Type: resolver.Symbol(here, f),
			Tag:  tag,
		}
	}
	return &Struct{
		Name:    name,
		Package: here,
		Fields:  fields,
	}
}

func WriteStruct(w io.Writer, here *tinypkg.Package, name string, s *Struct) error {
	// struct {
	// ..
	// }
	if name == "" {
		name = s.Name
	}
	fmt.Fprintf(w, "type %s struct {\n", name)
	defer fmt.Fprintln(w, "}")
	for _, field := range s.Fields {
		if field.Tag == "" {
			fmt.Fprintf(w, "\t%s %s\n", field.Name, tinypkg.ToRelativeTypeString(here, field.Type))
		} else {
			fmt.Fprintf(w, "\t%s %s `%s`\n", field.Name, tinypkg.ToRelativeTypeString(here, field.Type), field.Tag)
		}
	}
	return nil
}
