package resolve

import (
	"fmt"
	"reflect"

	reflectshape "github.com/podhmo/reflect-shape"
)

type StructFromShapeInfo struct {
	GroupedByKind map[Kind][]int
}

func StructFromShape(resolver *Resolver, fn reflectshape.Function) (reflectshape.Struct, StructFromShapeInfo, error) {
	fields := reflectshape.ShapeMap{}
	tags := make([]reflect.StructTag, 0, fn.Params.Len())
	metadata := make([]reflectshape.FieldMetadata, 0, fn.Params.Len())

	kindMap := map[Kind][]int{}

	for i, name := range fn.Params.Keys {
		p := fn.Params.Values[i]
		kind := resolver.DetectKind(p)

		switch kind {
		case KindIgnored, KindUnsupported, KindComponent:
			continue
		case KindData, KindPrimitive, KindPrimitivePointer:
			switch kind {
			case KindData:
				fields.Keys = append(fields.Keys, name)
				fields.Values = append(fields.Values, p)
				metadata = append(metadata, reflectshape.FieldMetadata{
					FieldName: name,
					Required:  true,
					Anonymous: true,
				})
				tags = append(tags, "")
				kindMap[kind] = append(kindMap[kind], i)
			case KindPrimitive:
				fields.Keys = append(fields.Keys, name)
				fields.Values = append(fields.Values, p)
				metadata = append(metadata, reflectshape.FieldMetadata{
					FieldName: name,
					Required:  true,
				})
				tags = append(tags, "")
				kindMap[kind] = append(kindMap[kind], i)
			case KindPrimitivePointer:
				fields.Keys = append(fields.Keys, name)
				fields.Values = append(fields.Values, p)
				metadata = append(metadata, reflectshape.FieldMetadata{
					FieldName: name,
					Required:  false,
				})
				tags = append(tags, "")
				kindMap[kind] = append(kindMap[kind], i)
			}
		default:
			return reflectshape.Struct{}, StructFromShapeInfo{}, fmt.Errorf("unsupported kind %v", kind)
		}
	}

	shape := reflectshape.Struct{
		Info: &reflectshape.Info{
			Name:    "", // not ref
			Kind:    reflectshape.Kind(reflect.Struct),
			Package: fn.Info.Package,
		},
		Fields:   fields,
		Tags:     tags,
		Metadata: metadata,
	}
	shape.ResetReflectType(reflect.PtrTo(fn.GetReflectType())) // xxx:
	return shape, StructFromShapeInfo{GroupedByKind: kindMap}, nil
}
