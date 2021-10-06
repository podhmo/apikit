package translate

import (
	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

var ErrNoImports = code.ErrNoImports
var DefaultConfig = code.DefaultConfig

type Translator struct {
	*resolve.Tracker
	Resolver    *resolve.Resolver
	Config      *code.Config
	providerVar *tinypkg.Var // TODO: from config
}

func NewTranslator(config *code.Config) *Translator {
	return &Translator{
		Tracker:  resolve.NewTracker(config.Resolver),
		Resolver: config.Resolver,
		Config:   config,
	}
}
