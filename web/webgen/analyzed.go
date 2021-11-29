package webgen

import (
	"fmt"
	"net/http"

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
		Ignored []*tinypkg.Var // rename: context.Context, etc...

		Provider *tinypkg.Var

		GetProviderFunc   *tinypkg.Func
		CreateHandlerFunc *tinypkg.Func // todo: fix
	}
	Names struct {
		Name           string
		ActionFunc     string // core action
		ActionFuncArgs []string
		QueryParams    string
		PathParams     string
	}
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

func ProviderModule(here *tinypkg.Package, resolver *resolve.Resolver, providerName string) (*resolve.Module, error) {
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
	return m, nil
}

func Analyze(
	here *tinypkg.Package,
	resolver *resolve.Resolver,
	tracker *resolve.Tracker,
	info *web.PathInfo, // todo: remove
	extraDefs []*resolve.Def,
	providerModule *resolve.Module,
) (*Analyzed, error) {
	// TODO: typed
	createHandlerFunc, err := providerModule.Type("createHandler")
	if err != nil {
		return nil, fmt.Errorf("in provider module, %w", err)
	}
	createHandlerFunc.Args[0].Name = "getProvider" // todo: remove
	getProviderFunc := createHandlerFunc.Args[0].Node.(*tinypkg.Func)
	getProviderFunc.Name = "getProvider" // todo: remove
	provider := &tinypkg.Var{Name: "provider", Node: getProviderFunc.Returns[0].Node}

	var componentBindings tinypkg.BindingList
	var pathBindings []*PathBinding
	var queryBindings []*QueryBinding
	var dataBindings []*DataBinding

	var ignored []*tinypkg.Var
	seen := map[reflectshape.Identity]bool{}
	def := info.Def

	argNames := make([]string, 0, len(def.Args))

	var subDepends []reflectshape.Function
	for _, x := range def.Args {
		shape := x.Shape
		sym := resolver.Symbol(here, shape)
		switch x.Kind {
		case resolve.KindIgnored: // e.g. context.Context
			seen[x.Shape.GetIdentity()] = true

			ignored = append(ignored, &tinypkg.Var{Name: x.Name, Node: sym})
			argNames = append(argNames, x.Name)
		case resolve.KindComponent:
			seen[x.Shape.GetIdentity()] = true

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
			argNames = append(argNames, x.Name)
		case resolve.KindPrimitive: // handle pathParams
			if v, ok := info.Vars[x.Name]; ok {
				pathBindings = append(pathBindings, &PathBinding{Name: x.Name, Var: v, Sym: resolver.Symbol(here, v.Shape)})
			}
			argNames = append(argNames, "pathParams."+x.Name)
		case resolve.KindPrimitivePointer: // handle query string
			queryBindings = append(queryBindings, &QueryBinding{Name: x.Name, Sym: resolver.Symbol(here, x.Shape)})
			argNames = append(argNames, "queryParams."+x.Name)
		case resolve.KindData: // handle request.Body
			dataBindings = append(dataBindings, &DataBinding{Name: x.Name, Sym: resolver.Symbol(here, x.Shape)})
			argNames = append(argNames, x.Name)
		default:
			argNames = append(argNames, x.Name)
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

	analyzed := &Analyzed{}

	analyzed.Bindings.Component = componentBindings
	analyzed.Bindings.Query = queryBindings
	analyzed.Bindings.Path = pathBindings
	analyzed.Bindings.Data = dataBindings

	analyzed.Vars.Ignored = ignored
	analyzed.Vars.Provider = provider
	analyzed.Vars.GetProviderFunc = getProviderFunc
	analyzed.Vars.CreateHandlerFunc = createHandlerFunc

	analyzed.Names.Name = info.Def.Name
	analyzed.Names.QueryParams = "queryParams"
	analyzed.Names.PathParams = "pathParams"
	analyzed.Names.ActionFunc = tinypkg.ToRelativeTypeString(here, info.Def.Symbol)
	analyzed.Names.ActionFuncArgs = argNames

	if len(componentBindings) > 0 || len(ignored) > 0 {
		if len(componentBindings)-len(extraDefs) == 0 {
			analyzed.Vars.Provider.Name = "_"
		}
	}
	return analyzed, nil
}
