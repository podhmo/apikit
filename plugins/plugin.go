package plugins

import (
	"context"
	"fmt"
	"reflect"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/emitgo"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

func NewDefaultConfig(emitter *emitgo.Emitter) *PluginConfig {
	config := code.DefaultConfig()
	return &PluginConfig{
		Config:   config,
		Emitter:  emitter,
		Resolver: config.Resolver,
	}
}

type PluginConfig struct {
	*code.Config

	Emitter  *emitgo.Emitter
	Resolver *resolve.Resolver
}

func (c *PluginConfig) ActivatePlugins(
	ctx context.Context, here *tinypkg.Package,
	plugins ...Plugin) error {
	pc := &PluginContext{Context: ctx, PluginConfig: c}
	for _, p := range plugins {
		if err := pc.IncludePlugin(here, p); err != nil {
			return fmt.Errorf("error in plugin %s: %w", nameOfPlugin(p), err)
		}
	}
	return nil
}

func nameOfPlugin(p Plugin) string {
	rt := reflect.TypeOf(p)
	return fmt.Sprintf("%s.%s", rt.PkgPath(), rt.Name())
}

type PluginContext struct {
	*PluginConfig
	Context context.Context
}

func (pc *PluginContext) IncludePlugin(here *tinypkg.Package, p Plugin) error {
	return p.IncludeMe(pc, here)
}

type Plugin interface {
	IncludeMe(pc *PluginContext, here *tinypkg.Package) error
}
