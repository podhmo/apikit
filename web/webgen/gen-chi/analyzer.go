package genchi

import (
	"fmt"
	"io"
	"net/http"

	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
	"github.com/podhmo/apikit/web/webgen"
)

type Analyzer struct {
	Resolver *resolve.Resolver
	Tracker  *resolve.Tracker

	ProviderModule *resolve.Module
	RuntimeModule  *resolve.Module
}

func newAnalyzer(g *Generator) (*Analyzer, error) {
	resolver := g.Resolver
	providerModule, err := ProviderModule(g.ProviderPkg, resolver, g.Config.ProviderName)
	if err != nil {
		return nil, err
	}
	runtimeModule, err := RuntimeModule(g.RuntimePkg, resolver)
	if err != nil {
		return nil, err
	}
	return &Analyzer{
		Resolver:       resolver,
		Tracker:        g.Tracker,
		ProviderModule: providerModule,
		RuntimeModule:  runtimeModule,
	}, nil
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

func ProviderModule(here *tinypkg.Package, resolver *resolve.Resolver, providerName string) (*resolve.Module, error) {
	type providerT interface{}
	var moduleSkeleton struct {
		T             providerT
		getProvider   func(*http.Request) (*http.Request, providerT, error)
		createHandler func(
			getProvider func(*http.Request) (*http.Request, providerT, error),
		) http.HandlerFunc
	}
	pm, err := resolver.PreModule(moduleSkeleton)
	if err != nil {
		return nil, fmt.Errorf("new provider pre-module: %w", err)
	}
	m, err := pm.NewModule(here, here.NewSymbol(providerName))
	if err != nil {
		return nil, fmt.Errorf("new provider module: %w", err)
	}
	return m, nil
}

type Analyzed struct {
	*webgen.Analyzed
	analyzer *Analyzer
}

func (a *Analyzer) Analyze(here *tinypkg.Package, node *web.WalkerNode) (*Analyzed, error) {
	def := a.Resolver.Def(node.Node.Value)
	a.Tracker.Track(def)

	pathinfo, err := web.ExtractPathInfo(node.Node.VariableNames, def)
	if err != nil {
		return nil, err
	}

	extraDeps := web.GetMetaData(node.Node).ExtraDependencies
	extraDefs := make([]*resolve.Def, len(extraDeps))
	for i, fn := range extraDeps {
		extraDef := a.Resolver.Def(fn)
		a.Tracker.Track(extraDef)
		extraDefs[i] = extraDef
	}

	analyzed, err := webgen.Analyze(
		here,
		a.Resolver, a.Tracker,
		pathinfo, extraDefs,
		a.ProviderModule,
	)
	if err != nil {
		return nil, err
	}
	return &Analyzed{
		Analyzed: analyzed,
		analyzer: a,
	}, nil
}
