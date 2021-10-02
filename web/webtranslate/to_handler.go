package webtranslate

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
	reflectshape "github.com/podhmo/reflect-shape"
)

func (t *Translator) RuntimeModule(here *tinypkg.Package) (*resolve.Module, error) {
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
	return m, nil
}

func (t *Translator) GetProviderModule(here *tinypkg.Package, providerName string) (*resolve.Module, error) {
	type providerT interface{}
	var moduleSkeleton struct {
		T           providerT
		getProvider func(*http.Request) (*http.Request, providerT, error)
	}
	pm, err := t.Resolver.PreModule(moduleSkeleton)
	if err != nil {
		return nil, fmt.Errorf("new provider pre-module: %w", err)
	}
	m, err := pm.NewModule(here, here.NewSymbol(providerName))
	if err != nil {
		return nil, fmt.Errorf("new provider module: %w", err)
	}
	return m, nil
}

func WriteHandlerFunc(w io.Writer, here *tinypkg.Package, resolver *resolve.Resolver, tracker *resolve.Tracker, info *web.PathInfo, providerModule *resolve.Module, runtimeModule *resolve.Module, name string) error {
	// TODO: typed
	getProviderFunc, err := providerModule.Type("getProvider")
	if err != nil {
		return fmt.Errorf("in provider module, %w", err)
	}
	handleResultFunc, err := runtimeModule.Symbol(here, "HandleResult")
	if err != nil {
		return fmt.Errorf("in runtime module, %w", err)
	}
	pathParamFunc, err := runtimeModule.Symbol(here, "PathParam")
	if err != nil {
		return fmt.Errorf("in runtime module, %w", err)
	}

	actionFunc := tinypkg.ToRelativeTypeString(here, info.Def.Symbol)

	args := []*tinypkg.Var{
		{Name: getProviderFunc.Name, Node: getProviderFunc},
	}
	provider := &tinypkg.Var{Name: "provider", Node: getProviderFunc.Returns[0].Node}
	if name == "" {
		name = info.Def.Name
	}

	var componentBindings []*tinypkg.Binding
	var pathBindings []*web.PathVar
	var ignored []*tinypkg.Var
	seen := map[reflectshape.Identity]bool{}
	def := info.Def

	argNames := make([]string, 0, len(def.Args))

	// TODO: handling path info
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
		case resolve.KindPrimitive:
			if v, ok := info.Vars[x.Name]; ok {
				pathBindings = append(pathBindings, v)
			}
		default:
			args = append(args, &tinypkg.Var{Name: x.Name, Node: sym})
		}
	}

	if len(pathBindings) != len(info.VarNames) {
		return fmt.Errorf("invalid path bindings, routing=%v, args=%v (in %s)", info.VarNames, pathBindings, info.Def.Symbol)
	}

	// TODO: handling ignore
	_ = ignored

	f := here.NewFunc(name, args, nil)
	return tinypkg.WriteFunc(w, here, "", f, func() error {
		fmt.Fprintf(w, "\treturn func(w http.ResponseWriter, req *http.Request) http.HandlerFunc{\n")
		defer fmt.Fprintln(w, "\t}")

		// <path name> := runtime.PathParam(req, "<path name>")
		if len(pathBindings) > 0 {
			for _, pathvar := range pathBindings {
				// TODO: type check
				fmt.Fprintf(w, "\t\t%s := %s(req, %q)\n", pathvar.Name, pathParamFunc, pathvar.Name)
			}
		}

		// var <component> <type>
		// {
		// 	<component> = <provider>.<method>()
		// }
		if len(componentBindings) > 0 {
			// handling req.Context
			fmt.Fprintf(w, "\t\treq, %s, err := %s(req)\n", provider.Name, getProviderFunc.Name)
			fmt.Fprintln(w, "\t\tif err != nil {")
			fmt.Fprintf(w, "\t\t\t%s(w, req, nil, err)\n", handleResultFunc)
			fmt.Fprintln(w, "\t\t\treturn")
			fmt.Fprintln(w, "\t\t}")
			indent := "\t\t"
			var returns []*tinypkg.Var

			// handling components
			zeroReturnsDefault := fmt.Sprintf("%s(w, req, nil, err); return", handleResultFunc)
			for _, binding := range componentBindings {
				binding.ZeroReturnsDefault = zeroReturnsDefault
				if err := binding.WriteWithCleanupAndError(w, here, indent, returns); err != nil {
					return err
				}
			}
		}

		// result, err := <action>(....)
		fmt.Fprintf(w, "\t\tresult, err := %s(%s)\n", actionFunc, strings.Join(argNames, ", "))

		// runtime.HandleResult(w, req, result, err)
		fmt.Fprintf(w, "\t\t%s(w, req, result, err)\n", handleResultFunc)
		return nil
	})
}
