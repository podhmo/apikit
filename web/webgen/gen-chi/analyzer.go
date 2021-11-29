package genchi

import (
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
