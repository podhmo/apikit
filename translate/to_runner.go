package translate

import (
	"fmt"
	"io"
	"strings"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	reflectshape "github.com/podhmo/reflect-shape"
)

// TODO: omit provider arguments

func (t *Translator) TranslateToRunner(here *tinypkg.Package, fn interface{}, name string, provider *tinypkg.Var) *code.Code {
	def := t.Resolver.Def(fn)
	if name == "" {
		name = def.Name
	}
	t.Tracker.Track(def)
	return &code.Code{
		Name: name,
		Here: here,
		// priority: code.PrioritySecond,
		Config: t.Config,
		ImportPackages: func() ([]*tinypkg.ImportedPackage, error) {
			if provider == nil {
				provider = t.providerVar
			}
			return collectImportsForRunner(here, t.Resolver, t.Tracker, def, provider)
		},
		EmitCode: func(w io.Writer) error {
			if provider == nil {
				provider = t.providerVar
			}
			return writeRunner(w, here, t.Resolver, t.Tracker, def, provider, name)
		},
	}
}

func collectImportsForRunner(here *tinypkg.Package, resolver *resolve.Resolver, tracker *resolve.Tracker, def *resolve.Def, provider *tinypkg.Var) ([]*tinypkg.ImportedPackage, error) {
	collector := tinypkg.NewImportCollector(here)
	use := collector.Collect
	for _, x := range def.Args {
		shape := tracker.ExtractComponentFactoryShape(x)
		sym := resolver.Symbol(here, shape)
		if err := tinypkg.Walk(sym, use); err != nil {
			return nil, err
		}
	}
	for _, x := range def.Returns {
		sym := resolver.Symbol(here, x.Shape)
		if err := tinypkg.Walk(sym, use); err != nil {
			return nil, err
		}
	}
	if err := tinypkg.Walk(provider, use); err != nil {
		return nil, err
	}
	return collector.Imports, nil
}

func writeRunner(w io.Writer, here *tinypkg.Package, resolver *resolve.Resolver, tracker *resolve.Tracker, def *resolve.Def, provider *tinypkg.Var, name string) error {
	var componentBindings []*tinypkg.Binding
	var ignored []*tinypkg.Var
	seen := map[reflectshape.Identity]bool{}

	argNames := make([]string, 0, len(def.Args))
	args := make([]*tinypkg.Var, 0, len(def.Args)+1)
	{
		sym := provider.Node.(*tinypkg.Symbol)
		args = append(args, &tinypkg.Var{Name: provider.Name, Node: here.Import(sym.Package).Lookup(sym)})
	}
	for _, x := range def.Args {
		argNames = append(argNames, x.Name)
		shape := x.Shape
		sym := resolver.Symbol(here, shape)
		switch x.Kind {
		case resolve.KindIgnored: // e.g. context.Context
			k := x.Shape.GetIdentity()
			if _, ok := seen[k]; !ok {
				seen[k] = true
				ignored = append(ignored, &tinypkg.Var{Name: x.Name, Node: sym})
			}
		case resolve.KindComponent:
			shape := tracker.ExtractComponentFactoryShape(x)

			if v, ok := shape.(reflectshape.Function); ok {
				for i, p := range v.Params.Values {
					switch resolve.DetectKind(p) {
					case resolve.KindIgnored: // e.g. context.Context
						k := p.GetIdentity()
						if _, ok := seen[k]; !ok {
							seen[k] = true
							xname := v.Params.Keys[i]
							sym := resolver.Symbol(here, p)
							ignored = append(ignored, &tinypkg.Var{Name: xname, Node: sym})
						}
					}
				}
			}

			sym := resolver.Symbol(here, shape)
			factory, ok := sym.(*tinypkg.Func)
			if !ok {
				// func() <component>
				factory = here.NewFunc("", nil, []*tinypkg.Var{{Node: sym}})
			}

			binding, err := tinypkg.NewBinding(x.Name, factory)
			if err != nil {
				return err
			}

			rt := x.Shape.GetReflectType() // not shape.GetRefectType()
			methodName := tracker.ExtractMethodName(rt, x.Name)
			binding.ProviderAlias = fmt.Sprintf("%s.%s", provider.Name, methodName)

			componentBindings = append(componentBindings, binding)
		default:
			args = append(args, &tinypkg.Var{Name: x.Name, Node: sym})
		}
	}

	if len(ignored) > 0 {
		args = append(ignored, args...)
	}

	returns := make([]*tinypkg.Var, 0, len(def.Returns))
	for _, x := range def.Returns {
		sym := resolver.Symbol(here, x.Shape)
		returns = append(returns, &tinypkg.Var{Node: sym}) // TODO: need using x.Name?
	}

	return tinypkg.WriteFunc(w, here, name, &tinypkg.Func{Args: args, Returns: returns},
		func() error {

			// var <component> <type>
			// {
			//   <component> = <provider>.<method>()
			// }
			if len(componentBindings) > 0 {
				indent := "\t"
				for _, binding := range componentBindings {
					if err := binding.WriteWithCleanupAndError(w, here, indent, returns); err != nil {
						return err
					}
				}
			}

			// return <inner function>(<args>...)
			fmt.Fprintf(w, "\treturn %s(%s)\n",
				tinypkg.ToRelativeTypeString(here, def.Symbol),
				strings.Join(argNames, ", "),
			)
			return nil
		},
	)
}
