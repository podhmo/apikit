package genlambda

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/translate"
)

type Translator struct {
	Resolver *resolve.Resolver
	Tracker  *resolve.Tracker
	Config   *code.Config

	ProviderModule *resolve.Module
	RuntimeModule  *resolve.Module

	internal *translate.Translator
}

func ProviderModule(here *tinypkg.Package, resolver *resolve.Resolver, providerName string) (*resolve.Module, error) {
	eventStructName := "Event" // todo: to arguments

	type providerT interface{}
	type eventT interface{}
	var moduleSkeleton struct {
		T             providerT
		EventT        eventT
		getProvider   func(context.Context) (context.Context, providerT, error)
		createHandler func(
			getProvider func(context.Context) (context.Context, providerT, error),
		) func(context.Context, eventT) (interface{}, error)
	}
	pm, err := resolver.PreModule(moduleSkeleton)
	if err != nil {
		return nil, fmt.Errorf("new provider pre-module: %w", err)
	}
	m, err := pm.NewModule(here, here.NewSymbol(providerName), here.NewSymbol(eventStructName))
	if err != nil {
		return nil, fmt.Errorf("new provider module: %w", err)
	}
	return m, nil
}

func RuntimeModule(here *tinypkg.Package, resolver *resolve.Resolver) (*resolve.Module, error) {
	var moduleSkeleton struct {
		PathParam                  func(*http.Request, string) string
		HandleResult               func(http.ResponseWriter, *http.Request, interface{}, error)
		CreateHandleResultFunction func(func(error) int) func(http.ResponseWriter, *http.Request, interface{}, error)

		BindPathParams func(dst interface{}, req *http.Request, keys ...string) error
		BindQuery      func(dst interface{}, req *http.Request) error
		BindBody       func(dst interface{}, src io.ReadCloser) error
		ValidateStruct func(ob interface{}) error
	}
	pm, err := resolver.PreModule(moduleSkeleton)
	if err != nil {
		return nil, fmt.Errorf("new runtime pre-module: %w", err)
	}
	m, err := pm.NewModule(here)
	if err != nil {
		return nil, fmt.Errorf("new runtime module: %w", err)
	}
	return m, nil
}
