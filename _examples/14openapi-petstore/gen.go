//go:build apikit
// +build apikit

// this code is generated by "apikit init"

package main

import (
	"context"
	"fmt"
	"log"
	"m/14openapi-petstore/action"
	"m/14openapi-petstore/design"
	"m/14openapi-petstore/myplugins/gendoc"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/web"
	genchi "github.com/podhmo/apikit/web/webgen/gen-chi"
)

// generate code: VERBOSE=1 go run gen.go

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func router() *web.Router {
	r := web.NewRouter()

	r.Group("", func(r *web.Router) {
		r.MetaData.Tags = []string{"pet"}

		r.Get("/pets", action.FindPets, web.WithTags("query"))
		r.Post("/pets", action.AddPet)
		r.Get("/pets/{id}", action.FindPetByID, web.WithTags("query"))
		r.Delete("/pets/{id}", action.DeletePet, web.WithDefaultStatusCode(204))
	})

	return r
}

func run() (err error) {
	emitter := emitgo.NewConfigFromRelativePath(action.AddPet, "..").NewEmitter()
	emitter.FilenamePrefix = "gen_" // generated file name is "gen_<name>.go"
	defer emitter.EmitWith(&err)

	c := genchi.DefaultConfig()
	// c.Override("logger", action.NewLogger) // register provider as func() (*log.Logger, error)

	g := c.New(emitter)
	r := router()
	if err := g.Generate(context.Background(), r, design.HTTPStatusOf); err != nil {
		return err
	}

	// generate openapi doc via custom plugin
	type defaultError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := g.IncludePlugin(g.RootPkg, gendoc.Options{
		OutputFile:   "docs/openapi.json",
		Handlers:     g.Handlers,
		DefaultError: defaultError{},
		Prepare: func(m *gendoc.Manager) {
			// customize information
			var doc *openapi3.T = m.Doc
			doc.Info.Title = "Swagger Petstore"
			doc.Info.Version = "1.0.0"
			doc.Info.Description = "A sample API that uses a petstore as an example to demostorate features in the OpenAPI 3.0 specification."
		},
	}); err != nil {
		return fmt.Errorf("on gendoc plugin: %w", err)
	}
	return nil

}
