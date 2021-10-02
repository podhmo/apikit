package webtranslate

import (
	"fmt"
	"reflect"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

var ErrNoImports = code.ErrNoImports

type Config struct {
	*code.Config
	Runtime *tinypkg.Package
}

func DefaultConfig() *Config {
	c := &Config{
		Config: code.DefaultConfig(),
	}
	return c
}

type Translator struct {
	Tracker     *resolve.Tracker
	Resolver    *resolve.Resolver
	Config      *Config
	providerVar *tinypkg.Var // TODO: from config
}

func NewTranslator(config *Config) *Translator {
	tracker := resolve.NewTracker()
	return &Translator{
		Tracker:  tracker,
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