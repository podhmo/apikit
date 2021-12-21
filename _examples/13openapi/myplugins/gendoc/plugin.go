package gendoc

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"reflect"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/emitfile"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/plugins"
	"github.com/podhmo/apikit/plugins/enum"
	"github.com/podhmo/apikit/resolve"
	genchi "github.com/podhmo/apikit/web/webgen/gen-chi"
	reflectopenapi "github.com/podhmo/reflect-openapi"
	reflectshape "github.com/podhmo/reflect-shape"
)

type Options struct {
	OutputFile   string
	Handlers     []genchi.Handler
	DefaultError interface{}
	Prepare      func(m *Manager)
}

func (o Options) IncludeMe(pc *plugins.PluginContext, here *tinypkg.Package) error {
	return IncludeMe(
		pc.Context,
		pc.Config, pc.Resolver, pc.Emitter,
		here,
		o.OutputFile,
		o.Handlers,
		o.DefaultError,
		o.Prepare,
	)
}

type Manager struct {
	*reflectopenapi.Manager
}

func (m *Manager) DefineEnum(value interface{}, values ...interface{}) {
	m.Visitor.VisitType(value, func(schema *openapi3.Schema) {
		schema.Enum = append([]interface{}{value}, values...)
	})
}
func (m *Manager) DefineEnumWithEnumSet(value interface{}, set enum.EnumSet) {
	m.Visitor.VisitType(value, func(schema *openapi3.Schema) {
		schema.Title = set.Name
		values := make([]interface{}, len(set.Enums))
		names := make([]string, len(set.Enums))
		descs := make([]string, len(set.Enums))
		hasDesc := false
		for i, x := range set.Enums {
			names[i] = x.Name
			values[i] = x.Value
			descs[i] = x.Description
			if x.Description != "" {
				hasDesc = true
			}
		}
		schema.Enum = values
		if schema.Extensions == nil {
			schema.Extensions = map[string]interface{}{}
		}
		schema.Extensions["x-enum-varnames"] = names
		if hasDesc {
			schema.Extensions["x-enum-descriptions"] = descs
		}
	})
}

func IncludeMe(
	ctx context.Context,
	config *code.Config, resolver *resolve.Resolver, emitter *emitgo.Emitter,
	here *tinypkg.Package,
	outputFile string,
	handlers []genchi.Handler,
	defaultError interface{},
	prepare func(m *Manager),
) error {
	if outputFile == "" {
		log.Println("output filename is empty, so saving at docs/openapi.json")
		outputFile = "docs/openapi.json"
	}

	rc := reflectopenapi.Config{
		SkipValidation: true,
		StrictSchema:   true,
		Extractor:      resolver.UnsafeShapeExtractor(),
		IsRequiredCheckFunction: func(tag reflect.StructTag) bool {
			v, _ := strconv.ParseBool(tag.Get("required"))
			return v
		},
		Selector: &MergeParamsSelector{resolver: resolver},
	}

	if defaultError != nil {
		rc.DefaultError = defaultError
	}

	doc, err := rc.BuildDoc(ctx, func(m *reflectopenapi.Manager) {
		if prepare != nil {
			prepare(&Manager{Manager: m})
		}

		for _, h := range handlers {
			analyzed := h.Analyzed
			metadata := h.MetaData

			m.RegisterFunc(h.RawFn).After(func(op *openapi3.Operation) {
				// handling tags
				{
					for p := &metadata; p.Parent != nil; p = p.Parent {
						if p.Tags != nil {
							op.Tags = append(op.Tags, p.Tags...)
						}
					}
				}

				// normalize the name of path-params
				{
					// e.g. articleID -> articleId
					for _, p := range op.Parameters {
						if p.Value.In == "path" {
							for _, binding := range analyzed.Bindings.Path {
								if binding.Name == p.Value.Name {
									p.Value.Name = binding.Var.Name
								}
							}
						}
					}
				}

				// handling default status
				{
					if code := metadata.DefaultStatusCode; code != 0 && op.Responses != nil {
						if v200, ok := op.Responses["200"]; ok {
							op.Responses[strconv.Itoa(code)] = v200
							delete(op.Responses, "200")
						}
					}
				}

				m.Doc.AddOperation(metadata.Path, metadata.Method, op)
			})
		}
	})
	if err != nil {
		log.Printf("WARNING: generate doc is failured %+v", err)
		return nil
	}

	emitter.FileEmitter.Register(outputFile, emitfile.EmitFunc(func(w io.Writer) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(doc)
	}))
	return nil
}

// selector ////////////////////////////////////////
//
// func(data Data, xxID string) { ... }
//
// to
//
// struct {
//  Data
//  xxID string `openapi:"path"`
// }
//

// MergeParamsSelector is the selector with merging function arguments as single struct
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
	shape, info, err := resolve.StructFromShape(s.resolver, fn)
	if err != nil {
		panic(err) // xxx
	}

	// add openapi tag
	if indices, ok := info.GroupedByKind[resolve.KindPrimitive]; ok {
		for _, i := range indices {
			shape.Tags[i] = reflect.StructTag(string(shape.Tags[i]) + ` openapi:"path"`)
		}
	}
	if indices, ok := info.GroupedByKind[resolve.KindPrimitivePointer]; ok {
		for _, i := range indices {
			shape.Tags[i] = reflect.StructTag(string(shape.Tags[i]) + ` openapi:"query"`)
		}
	}
	return shape
}
