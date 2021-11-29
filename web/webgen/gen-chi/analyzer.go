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

	ProviderModule *providerModule
	runtimeModule  *runtimeModule
}

func newAnalyzer(g *Generator) (*Analyzer, error) {
	resolver := g.Resolver
	providerModule, err := ProviderModule(g.ProviderPkg, resolver, g.Config.ProviderName)
	if err != nil {
		return nil, fmt.Errorf("in provider module: %w", err)
	}
	runtimeModule, err := RuntimeModule(g.RuntimePkg, resolver)
	if err != nil {
		return nil, fmt.Errorf("in runtime module: %w", err)
	}
	return &Analyzer{
		Resolver:       resolver,
		Tracker:        g.Tracker,
		ProviderModule: providerModule,
		runtimeModule:  runtimeModule,
	}, nil
}

func RuntimeModule(here *tinypkg.Package, resolver *resolve.Resolver) (*runtimeModule, error) {
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
	return &runtimeModule{Module: m, Initialized: false}, nil
}

func (m *runtimeModule) Imported(here *tinypkg.Package) (*runtimeModule, error) {
	handleResultFunc, err := m.Symbol(here, "HandleResult")
	if err != nil {
		return nil, err
	}
	bindPathParamsFunc, err := m.Symbol(here, "BindPathParams")
	if err != nil {
		return nil, err
	}
	bindQueryFunc, err := m.Symbol(here, "BindQuery")
	if err != nil {
		return nil, err
	}
	bindBodyFunc, err := m.Symbol(here, "BindBody")
	if err != nil {
		return nil, err
	}
	validateStructFunc, err := m.Symbol(here, "ValidateStruct")
	if err != nil {
		return nil, err
	}
	createHandleResultFunc, err := m.Symbol(here, "CreateHandleResultFunction")
	if err != nil {
		return nil, err
	}

	mm := &runtimeModule{Module: m.Module, Initialized: true}
	mm.Symbols.HandleResult = handleResultFunc
	mm.Symbols.CreateHandleResult = createHandleResultFunc
	mm.Symbols.BindPathParams = bindPathParamsFunc
	mm.Symbols.BindQuery = bindQueryFunc
	mm.Symbols.BindBody = bindBodyFunc
	mm.Symbols.ValidateStruct = validateStructFunc
	return mm, nil
}

type runtimeModule struct {
	*resolve.Module
	Initialized bool
	Symbols     struct {
		HandleResult       *tinypkg.ImportedSymbol
		CreateHandleResult *tinypkg.ImportedSymbol
		BindPathParams     *tinypkg.ImportedSymbol
		BindQuery          *tinypkg.ImportedSymbol
		BindBody           *tinypkg.ImportedSymbol
		ValidateStruct     *tinypkg.ImportedSymbol
	}
}

func ProviderModule(here *tinypkg.Package, resolver *resolve.Resolver, providerName string) (*providerModule, error) {
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

	// TODO: detect name automatically
	createHandlerFunc, err := m.Type("createHandler")
	if err != nil {
		return nil, err
	}
	createHandlerFunc.Args[0].Name = "getProvider" // todo: remove
	getProviderFunc := createHandlerFunc.Args[0].Node.(*tinypkg.Func)
	getProviderFunc.Name = "getProvider" // todo: remove
	provider := &tinypkg.Var{Name: "provider", Node: getProviderFunc.Returns[0].Node}

	mm := &providerModule{Module: m}
	mm.Funcs.CreateHandler = createHandlerFunc
	mm.Funcs.GetProvider = getProviderFunc
	mm.Vars.Provider = provider
	return mm, nil
}

type providerModule struct {
	*resolve.Module

	Funcs struct {
		CreateHandler *tinypkg.Func
		GetProvider   *tinypkg.Func
	}
	Vars struct {
		Provider *tinypkg.Var
	}
}

type Analyzed struct {
	*webgen.Analyzed
	RuntimeModule  *runtimeModule
	ProviderModule *providerModule
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
		a.ProviderModule.Vars.Provider,
	)
	if err != nil {
		return nil, err
	}

	importedRuntimeModule, err := a.runtimeModule.Imported(here) // import m/runtime
	return &Analyzed{
		Analyzed:       analyzed,
		RuntimeModule:  importedRuntimeModule,
		ProviderModule: a.ProviderModule,
	}, nil
}
