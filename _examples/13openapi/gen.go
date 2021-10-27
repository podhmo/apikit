// +build apikit

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"

	"m/13openapi/design"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/podhmo/apikit/pkg/emitfile"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
	genchi "github.com/podhmo/apikit/web/webgen/gen-chi"
	reflectopenapi "github.com/podhmo/reflect-openapi"
	reflectshape "github.com/podhmo/reflect-shape"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

// TODO: set error handler (500-handler)
// TODO: set 404-handler

func run() (err error) {
	emitter := emitgo.NewConfigFromRelativePath(design.ListArticle, "..").NewEmitter()
	defer emitter.EmitWith(&err)

	c := genchi.DefaultConfig()
	c.Override("db", design.NewDB)

	////////////////////////////////////////
	rc := reflectopenapi.Config{
		SkipValidation: true,
		StrictSchema:   true,
		Extractor:      c.Resolver.UnsafeShapeExtractor(),
		IsRequiredCheckFunction: func(tag reflect.StructTag) bool {
			v, _ := strconv.ParseBool(tag.Get("required"))
			return v
		},
		Selector: &MergeParamsSelector{resolver: c.Resolver},
	}
	////////////////////////////////////////

	var r *web.Router

	// TODO: use shape
	// TODO: share shape-extractor
	doc, err := rc.BuildDoc(context.Background(), func(m *reflectopenapi.Manager) {
		r = web.NewRouter()
		r.Group("/articles", func(r *web.Router) {
			// TODO: set tag
			r.Get("/", design.ListArticle, WithOpenAPIOperation(m))
			r.Get("/{articleId}", design.GetArticle, WithOpenAPIOperation(m))
			r.Post("/{articleId}/comments", design.PostArticleComment, WithOpenAPIOperation(m))
		})
	})
	if err != nil {
		return err
	}

	emitter.FileEmitter.Register("docs/openapi.json", emitfile.EmitFunc(func(w io.Writer) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(doc)
	}))

	g := c.New(emitter)
	if err := g.Generate(
		context.Background(),
		r,
		design.HTTPStatusOf,
	); err != nil {
		return err
	}
	return nil
}

func WithOpenAPIOperation(m *reflectopenapi.Manager) web.RoutingOption {
	return func(node *web.Node, metadata *web.MetaData) {
		m.RegisterFunc(node.Value).After(func(op *openapi3.Operation) {
			m.Doc.AddOperation(metadata.Path, metadata.Method, op)
		})
	}
}

type MergeParamsSelector struct {
	resolver *resolve.Resolver
	reflectopenapi.FirstParamOutputSelector
}

func (s *MergeParamsSelector) useArglist() {
}
func (s *MergeParamsSelector) SelectInput(fn reflectshape.Function) reflectshape.Shape {
	if len(fn.Params.Values) == 0 {
		return nil
	}
	fields := reflectshape.ShapeMap{}
	tags := make([]reflect.StructTag, 0, fn.Params.Len())
	metadata := make([]reflectshape.FieldMetadata, 0, fn.Params.Len())
	for i, name := range fn.Params.Keys {
		p := fn.Params.Values[i]
		switch kind := s.resolver.DetectKind(p); kind {
		case resolve.KindIgnored, resolve.KindUnsupported, resolve.KindComponent:
			continue
		case resolve.KindData, resolve.KindPrimitive, resolve.KindPrimitivePointer:
			switch kind {
			case resolve.KindData:
				s := p.(reflectshape.Struct)
				fields.Keys = append(fields.Keys, s.Fields.Keys...)
				fields.Values = append(fields.Values, s.Fields.Values...)
				metadata = append(metadata, s.Metadata...)
				tags = append(tags, s.Tags...)
			case resolve.KindPrimitive:
				fields.Keys = append(fields.Keys, name)
				fields.Values = append(fields.Values, p)
				metadata = append(metadata, reflectshape.FieldMetadata{
					FieldName: name,
					Required:  true,
				})
				tags = append(tags, reflect.StructTag(`openapi:"path"`)) // todo: see path param (e.g. articleId)
			case resolve.KindPrimitivePointer:
				fields.Keys = append(fields.Keys, name)
				fields.Values = append(fields.Values, p)
				metadata = append(metadata, reflectshape.FieldMetadata{
					FieldName: name,
					Required:  false,
				})
				tags = append(tags, reflect.StructTag(`openapi:"query"`))
			}
		default:
			panic(fmt.Sprintf("unsupported kind %v", kind))
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
	retval.ResetReflectType(reflect.PtrTo(fn.GetReflectType()))
	return retval
}
