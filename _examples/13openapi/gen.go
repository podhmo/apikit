//go:build apikit
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
	ctx := context.Background()

	emitter := emitgo.NewConfigFromRelativePath(design.ListArticle, "..").NewEmitter()
	defer emitter.EmitWith(&err)

	c := genchi.DefaultConfig()
	c.Override("db", design.NewDB)

	r := web.NewRouter()
	r.Group("/articles", func(r *web.Router) {
		// TODO: set tag
		r.Get("/", design.ListArticle)
		r.Get("/{articleId}", design.GetArticle)
		r.Post("/{articleId}/comments", design.PostArticleComment)
	})

	////////////////////////////////////////
	// TODO: share shape-extractor
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

	m, generateDoc, err := rc.NewManager()
	if err != nil {
		return err
	}
	c.OnAnalyzed = func(analyzed *genchi.Analyzed, h genchi.Handler) error {
		metadata := h.MetaData
		fn := h.RawFn
		m.RegisterFunc(fn).After(func(op *openapi3.Operation) {
			for _, p := range op.Parameters {
				if p.Value.In == "path" {
					for _, binding := range analyzed.Bindings.Path {
						if binding.Name == p.Value.Name {
							p.Value.Name = binding.Var.Name // e.g. articleID -> articleId
							fmt.Println("@", p.Value.Name, binding.Var.Name)
						}
					}
				}
			}
			m.Doc.AddOperation(metadata.Path, metadata.Method, op)
		})
		return nil
	}
	emitter.FileEmitter.Register("docs/openapi.json", emitfile.EmitFunc(func(w io.Writer) error {
		if err := generateDoc(ctx); err != nil {
			return fmt.Errorf("generate openAPI doc: %w", err)
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(m.Doc) // xxx
	}))
	////////////////////////////////////////

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
