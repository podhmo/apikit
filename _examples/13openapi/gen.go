// +build apikit

package main

import (
	"context"
	"encoding/json"
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
	shape, info, err := resolve.StructFromShape(s.resolver, fn, resolve.StructFromShapeOptions{SquashEmbedded: false})
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
