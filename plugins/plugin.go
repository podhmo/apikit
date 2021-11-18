package plugins

import (
	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

func NewDefaultPluginContext(emitter *emitgo.Emitter) *PluginContext {
	config := code.DefaultConfig()
	return &PluginContext{
		Config:   config,
		Emitter:  emitter,
		Resolver: config.Resolver,
	}
}

type PluginContext struct {
	*code.Config

	Emitter  *emitgo.Emitter
	Resolver *resolve.Resolver
}

func (pc *PluginContext) IncludePlugin(here *tinypkg.Package, p Plugin) error {
	return p.IncludeMe(pc, here)
}

type Plugin interface {
	IncludeMe(pc *PluginContext, here *tinypkg.Package) error
}
