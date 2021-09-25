package translate

import (
	"fmt"
	"reflect"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

var ErrNoImports = code.ErrNoImports
var DefaultConfig = code.DefaultConfig

type Translator struct {
	Tracker     *Tracker
	Resolver    *resolve.Resolver
	Config      *code.Config
	providerVar *tinypkg.Var // TODO: from config
}

func NewTranslator(config *code.Config) *Translator {
	tracker := NewTracker()
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
