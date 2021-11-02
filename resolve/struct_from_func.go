package resolve

import (
	"fmt"
	"reflect"

	reflectshape "github.com/podhmo/reflect-shape"
)

type StructFromShapeOptions struct {
	SquashEmbedded bool
}

func StructFromShape(resolver *Resolver, fn reflectshape.Function, options StructFromShapeOptions) (reflectshape.Struct, error) {
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
				if options.SquashEmbedded {
					fields.Keys = append(fields.Keys, s.Fields.Keys...)
					fields.Values = append(fields.Values, s.Fields.Values...)
					metadata = append(metadata, s.Metadata...)
					tags = append(tags, s.Tags...)
				} else {
					fields.Keys = append(fields.Keys, name)
					fields.Values = append(fields.Values, p)
					metadata = append(metadata, reflectshape.FieldMetadata{
						FieldName: name,
						Required:  true,
						Anonymous: true,
					})
					tags = append(tags, "")
				}
			case KindPrimitive:
				fields.Keys = append(fields.Keys, name)
				fields.Values = append(fields.Values, p)
				metadata = append(metadata, reflectshape.FieldMetadata{
					FieldName: name,
					Required:  true,
				})
				tags = append(tags, "")
			case KindPrimitivePointer:
				fields.Keys = append(fields.Keys, name)
				fields.Values = append(fields.Values, p)
				metadata = append(metadata, reflectshape.FieldMetadata{
					FieldName: name,
					Required:  false,
				})
				tags = append(tags, "")
			}
		default:
			return reflectshape.Struct{}, fmt.Errorf("unsupported kind %v", kind)
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
