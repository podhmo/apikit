package webgen

import (
	"fmt"

	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
	reflectshape "github.com/podhmo/reflect-shape"
)

type Analyzed struct {
	Bindings struct {
		Path      []*PathBinding
		Query     []*QueryBinding
		Component tinypkg.BindingList
		Data      []*DataBinding
	}
	Vars struct {
		Ignored  []*tinypkg.Var // rename: context.Context, etc...
		Provider *tinypkg.Var
	}
	Names struct {
		ActionFunc     string // core action
		ActionFuncArgs []fmt.Stringer
		QueryParams    *string
		PathParams     *string
	}

	Name      string
	PathInfo  *web.PathInfo
	extraDefs []*resolve.Def
	resolver  *resolve.Resolver
	tracker   *resolve.Tracker
}

func (a *Analyzed) CollectImports(collector *tinypkg.ImportCollector) error {
	def := a.PathInfo.Def
	if err := a.collectImportsFromDef(collector, def); err != nil {
		return err
	}

	if len(a.extraDefs) > 0 {
		for _, extraDef := range a.extraDefs {
			if err := a.collectImportsFromDef(collector, extraDef); err != nil {
				return err
			}
		}
	}

	return a.Vars.Provider.OnWalk(collector.Collect)
}

func (a *Analyzed) collectImportsFromDef(collector *tinypkg.ImportCollector, def *resolve.Def) error {
	here := collector.Here
	use := collector.Collect

	resolver := a.resolver

	for _, x := range def.Args {
		sym := resolver.Symbol(here, x.Shape)
		if err := tinypkg.Walk(sym, use); err != nil {
			return fmt.Errorf("on walk args %s: %w", sym, err)
		}
	}
	for _, x := range def.Returns {
		sym := resolver.Symbol(here, x.Shape)
		if err := tinypkg.Walk(sym, use); err != nil {
			return fmt.Errorf("on walk returns %s: %w", sym, err)
		}
	}
	if err := use(def.Symbol); err != nil {
		return fmt.Errorf("on self %s: %w", def.Symbol, err)
	}
	return nil
}

type PathBinding struct {
	Name string // go's name
	Var  *web.PathVar
	Sym  tinypkg.Node
}
type QueryBinding struct {
	Name string
	Sym  tinypkg.Node
}
type DataBinding struct {
	Name string
	Sym  tinypkg.Node
}

func Analyze(
	here *tinypkg.Package,
	resolver *resolve.Resolver,
	tracker *resolve.Tracker,
	info *web.PathInfo, // todo: remove
	extraDefs []*resolve.Def,
	provider *tinypkg.Var,
) (*Analyzed, error) {
	pathParamName := "pathParams"
	queryParamName := "queryParams"

	var componentBindings tinypkg.BindingList
	var pathBindings []*PathBinding
	var queryBindings []*QueryBinding
	var dataBindings []*DataBinding

	var ignored []*tinypkg.Var
	seen := map[reflectshape.Identity]bool{}
	def := info.Def

	argNames := make([]fmt.Stringer, 0, len(def.Args))

	var subDepends []reflectshape.Function
	for _, x := range def.Args {
		shape := x.Shape
		sym := resolver.Symbol(here, shape)
		switch x.Kind {
		case resolve.KindIgnored: // e.g. context.Context
			seen[x.Shape.GetIdentity()] = true

			ignored = append(ignored, &tinypkg.Var{Name: x.Name, Node: sym})
			argNames = append(argNames, tinypkg.Name(x.Name))
		case resolve.KindComponent:
			seen[x.Shape.GetIdentity()] = true

			shape := tracker.ExtractComponentFactoryShape(x)
			if v, ok := shape.(reflectshape.Function); ok {
				subDepends = append(subDepends, v)
			}

			sym := resolver.Symbol(here, shape)
			factory, ok := sym.(*tinypkg.Func)
			if !ok {
				// func() <component.
				factory = here.NewFunc("", nil, []*tinypkg.Var{{Node: sym}})
			}

			binding, err := tinypkg.NewBinding(x.Name, factory)
			if err != nil {
				return nil, fmt.Errorf("new binding name=%q : %w", x.Name, err)
			}

			rt := x.Shape.GetReflectType() // not shape.GetRefectType()
			methodName := tracker.ExtractMethodName(rt, x.Name)
			binding.ProviderAlias = fmt.Sprintf("%s.%s", provider.Name, methodName)

			componentBindings = append(componentBindings, binding)
			argNames = append(argNames, tinypkg.Name(x.Name))
		case resolve.KindPrimitive: // handle pathParams
			if v, ok := info.Vars[x.Name]; ok {
				pathBindings = append(pathBindings, &PathBinding{Name: x.Name, Var: v, Sym: resolver.Symbol(here, v.Shape)})
			}
			argNames = append(argNames, tinypkg.LazyPrefixName{Prefix: &pathParamName, Value: "." + x.Name})
		case resolve.KindPrimitivePointer: // handle query string
			queryBindings = append(queryBindings, &QueryBinding{Name: x.Name, Sym: resolver.Symbol(here, x.Shape)})
			argNames = append(argNames, tinypkg.LazyPrefixName{Prefix: &queryParamName, Value: "." + x.Name})
		case resolve.KindData: // handle request.Body
			dataBindings = append(dataBindings, &DataBinding{Name: x.Name, Sym: resolver.Symbol(here, x.Shape)})
			argNames = append(argNames, tinypkg.Name(x.Name))
		default:
			argNames = append(argNames, tinypkg.Name(x.Name))
		}
	}

	if len(extraDefs) > 0 {
		for i, extraDef := range extraDefs {
			for _, x := range extraDef.Args {
				k := x.Shape.GetIdentity()
				if _, ok := seen[k]; ok {
					continue
				}

				shape := x.Shape
				sym := resolver.Symbol(here, shape)
				switch x.Kind {
				case resolve.KindIgnored: // e.g. context.Context
					seen[k] = true
					ignored = append(ignored, &tinypkg.Var{Name: x.Name, Node: sym})
				case resolve.KindComponent:
					seen[k] = true

					shape := tracker.ExtractComponentFactoryShape(x)
					if v, ok := shape.(reflectshape.Function); ok {
						subDepends = append(subDepends, v)
					}

					sym := resolver.Symbol(here, shape)
					factory, ok := sym.(*tinypkg.Func)
					if !ok {
						// func() <component>
						factory = here.NewFunc("", nil, []*tinypkg.Var{{Node: sym}})
					}

					binding, err := tinypkg.NewBinding(x.Name, factory)
					if err != nil {
						return nil, fmt.Errorf("new binding name=%q : %w", x.Name, err)
					}

					rt := x.Shape.GetReflectType() // not shape.GetRefectType()
					methodName := tracker.ExtractMethodName(rt, x.Name)
					binding.ProviderAlias = fmt.Sprintf("%s.%s", provider.Name, methodName)

					componentBindings = append(componentBindings, binding)
				default:
					// noop
				}
			}
			extraBinding, err := tinypkg.NewBinding(
				fmt.Sprintf("extra%d", i),
				resolver.Symbol(here, extraDef.Shape).(*tinypkg.Func),
			)
			if err != nil {
				return nil, fmt.Errorf("new binding name=%q : %w", fmt.Sprintf("extra%d", i), err)
			}
			componentBindings = append(componentBindings, extraBinding)
		}
	}
	if len(subDepends) > 0 {
		for _, v := range subDepends {
			for i, name := range v.Params.Keys {
				subShape := v.Params.Values[i]
				k := subShape.GetIdentity() // todo: with name
				if _, ok := seen[k]; ok {
					continue
				}
				seen[k] = true

				switch kind := resolver.DetectKind(subShape); kind {
				case resolve.KindIgnored: // e.g. context.Context
					ignored = append(ignored, &tinypkg.Var{
						Name: name,
						Node: resolver.Symbol(here, subShape),
					})
				case resolve.KindComponent:
					sym := resolver.Symbol(here, tracker.ExtractComponentFactoryShape(resolve.Item{
						Kind:  kind,
						Name:  name,
						Shape: subShape,
					}))
					factory, ok := sym.(*tinypkg.Func)
					if !ok {
						// func() <component>
						factory = here.NewFunc("", nil, []*tinypkg.Var{{Node: sym}})
					}

					binding, err := tinypkg.NewBinding(name, factory)
					if err != nil {

						return nil, fmt.Errorf("new binding name=%q : %w", name, err)
					}

					rt := subShape.GetReflectType() // not shape.GetRefectType()
					methodName := tracker.ExtractMethodName(rt, name)
					binding.ProviderAlias = fmt.Sprintf("%s.%s", provider.Name, methodName)

					componentBindings = append(componentBindings, binding)
				}
			}
		}
	}

	analyzed := &Analyzed{
		resolver:  resolver,
		tracker:   tracker,
		Name:      info.Def.Name,
		PathInfo:  info,
		extraDefs: extraDefs,
	}

	analyzed.Bindings.Component = componentBindings
	analyzed.Bindings.Query = queryBindings
	analyzed.Bindings.Path = pathBindings
	analyzed.Bindings.Data = dataBindings

	analyzed.Vars.Ignored = ignored
	analyzed.Vars.Provider = provider

	analyzed.Names.QueryParams = &queryParamName
	analyzed.Names.PathParams = &pathParamName
	analyzed.Names.ActionFunc = tinypkg.ToRelativeTypeString(here, info.Def.Symbol)
	analyzed.Names.ActionFuncArgs = argNames

	// if provider is not used, changes to "_"
	if len(componentBindings) > 0 || len(ignored) > 0 {
		if len(componentBindings)-len(extraDefs) == 0 {
			analyzed.Vars.Provider = &tinypkg.Var{Name: "_", Node: analyzed.Vars.Provider.Node}
		}
	}
	return analyzed, nil
}
