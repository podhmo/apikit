package webtranslate

import (
	"fmt"
	"io"

	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
	reflectshape "github.com/podhmo/reflect-shape"
)

func WriteHandlerFunc(w io.Writer, here *tinypkg.Package, resolver *resolve.Resolver, tracker *resolve.Tracker, info *web.PathInfo, runtime *tinypkg.Package, getProviderFunc *tinypkg.Func, name string) error {
	args := []*tinypkg.Var{
		{Name: getProviderFunc.Name, Node: getProviderFunc},
	}
	provider := &tinypkg.Var{Name: "provider", Node: getProviderFunc.Returns[0].Node}
	if name == "" {
		name = info.Def.Name
	}

	var componentBindings []*tinypkg.Binding
	var ignored []*tinypkg.Var
	seen := map[reflectshape.Identity]bool{}
	def := info.Def
	// TODO: handling path info
	for _, x := range def.Args {
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

	// TODO: handling ignore
	_ = ignored

	f := here.NewFunc(name, args, nil)
	return tinypkg.WriteFunc(w, here, "", f, func() error {
		// TODO: import http
		fmt.Fprintf(w, "\treturn func(w http.ResponseWriter, req *http.Request) http.HandlerFunc{\n")
		if len(componentBindings) > 0 {
			indent := "\t\t"
			var returns []*tinypkg.Var

			binding, err := tinypkg.NewBinding(provider.Name, getProviderFunc)
			if err != nil {
				return err
			}
			if err := binding.WriteWithCleanupAndError(w, here, indent, returns); err != nil {
				return err
			}

			for _, binding := range componentBindings {
				if err := binding.WriteWithCleanupAndError(w, here, indent, returns); err != nil {
					return err
				}
			}
		}

		fmt.Fprintf(w, "\t\tresult, err := %s()\n",
			tinypkg.ToRelativeTypeString(here, info.Def.Symbol),
		)
		fmt.Fprintf(w, "\t\treturn %s(w, req, result, err)\n", tinypkg.ToRelativeTypeString(here, runtime.NewSymbol("HandleResult")))
		defer fmt.Fprintln(w, "\t}")
		return nil
	})
}

// type GetProviderFunc func(*http.Request) (*http.Request, Provider, error)

// func ListMessageHandler(getProvider GetProviderFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, req *http.Request) {
// 		req, provider, err := getProvider(req)
// 		if err != nil {
// 			webruntime.HandleResult(w, req, nil, err)
// 			return
// 		}

// 		var db *DB
// 		{
// 			db = provider.DB()
// 		}

// 		result, err := ListMesesage(db)
// 		webruntime.HandleResult(w, req, result, err)
// 	}
// }
