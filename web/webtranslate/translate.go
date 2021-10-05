package webtranslate

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

var ErrNoImports = code.ErrNoImports

type Config struct {
	*code.Config
	RootPkg *tinypkg.Package

	RuntimePkg   *tinypkg.Package
	ProviderPkg  *tinypkg.Package
	ProviderName string
}

func DefaultConfig() *Config {
	c := code.DefaultConfig()
	return &Config{
		Config:       c,
		ProviderName: "Provider",
		RootPkg:      c.Resolver.NewPackage("", ""),
	}
}

type Translator struct {
	Tracker  *resolve.Tracker
	Resolver *resolve.Resolver
	Config   *Config

	runtimeModule  *resolve.Module
	providerModule *resolve.Module
}

func NewTranslator(config *Config) *Translator {
	return &Translator{
		Tracker:  resolve.NewTracker(),
		Resolver: config.Resolver,
		Config:   config,
	}
}

func (t *Translator) Override(name string, providerFunc interface{}) (prev *resolve.Def, err error) {
	rt := reflect.TypeOf(providerFunc)
	if rt.Kind() != reflect.Func {
		return nil, fmt.Errorf("unexpected providerFunc, only function %v", rt)
	}
	return t.Tracker.Override(rt.Out(0), name, t.Resolver.Def(providerFunc)), nil
}

func (t *Translator) RuntimeModule() (*resolve.Module, error) {
	if t.Config.RuntimePkg == nil {
		t.Config.RuntimePkg = t.Config.RootPkg.Relative("runtime", "")
	}
	here := t.Config.RuntimePkg
	if t.runtimeModule != nil {
		if t.runtimeModule.Here != here {
			return nil, fmt.Errorf("conflict package, %v != %v", t.runtimeModule.Here, here)
		}
		return t.runtimeModule, nil
	}

	var moduleSkeleton struct {
		PathParam    func(*http.Request, string) string
		HandleResult func(http.ResponseWriter, *http.Request, interface{}, error)
	}
	pm, err := t.Resolver.PreModule(moduleSkeleton)
	if err != nil {
		return nil, fmt.Errorf("new runtime pre-module: %w", err)
	}
	m, err := pm.NewModule(here)
	if err != nil {
		return nil, fmt.Errorf("new runtime module: %w", err)
	}
	t.runtimeModule = m
	return m, nil
}

func (t *Translator) ProviderModule() (*resolve.Module, error) {
	providerName := t.Config.ProviderName
	if t.Config.ProviderPkg == nil {
		t.Config.ProviderPkg = t.Config.RootPkg.Relative("provider", "")
	}
	here := t.Config.ProviderPkg
	if t.providerModule != nil {
		if t.providerModule.Here != here {
			return nil, fmt.Errorf("conflict package, %v != %v", t.providerModule.Here, here)
		}
		return t.providerModule, nil
	}

	type providerT interface{}
	var moduleSkeleton struct {
		T             providerT
		getProvider   func(*http.Request) (*http.Request, providerT, error)
		createHandler func(
			getProvider func(*http.Request) (*http.Request, providerT, error),
		) http.HandlerFunc
	}
	pm, err := t.Resolver.PreModule(moduleSkeleton)
	if err != nil {
		return nil, fmt.Errorf("new provider pre-module: %w", err)
	}
	m, err := pm.NewModule(here, here.NewSymbol(providerName))
	if err != nil {
		return nil, fmt.Errorf("new provider module: %w", err)
	}
	t.providerModule = m
	return m, nil
}
