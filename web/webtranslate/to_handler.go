package webtranslate

import (
	"fmt"
	"io"
	"strings"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
	reflectshape "github.com/podhmo/reflect-shape"
)

func (t *Translator) TranslateToHandler(here *tinypkg.Package, node *web.WalkerNode, name string) *code.CodeEmitter {
	def := t.Resolver.Def(node.Node.Value)
	if name == "" {
		name = def.Name
	}
	t.Tracker.Track(def)

	c := &code.Code{
		Name: name,
		Here: here,
		// priority: code.PrioritySecond,
		Config: t.Config.Config,
		ImportPackages: func(collector *tinypkg.ImportCollector) error {
			// todo: support provider *tinypkg.Var
			if err := collectImportsForHandler(collector, t.Resolver, t.Tracker, def); err != nil {
				return err
			}
			// todo: remove if unused
			if err := collector.Add(here.Import(t.Config.RuntimePkg)); err != nil {
				return err
			}
			if err := collector.Add(here.Import(t.Resolver.NewPackage("net/http", ""))); err != nil {
				return err
			}
			return nil
		},
		EmitCode: func(w io.Writer) error {
			pathinfo, err := web.ExtractPathInfo(node.Node.VariableNames, def)
			if err != nil {
				return err
			}
			providerModule, err := t.ProviderModule()
			if err != nil {
				return err
			}
			runtimeModule, err := t.RuntimeModule()
			if err != nil {
				return err
			}
			return WriteHandlerFunc(w, here, t.Resolver, t.Tracker, pathinfo, providerModule, runtimeModule, name)
		},
	}
	return &code.CodeEmitter{Code: c}
}

func collectImportsForHandler(collector *tinypkg.ImportCollector, resolver *resolve.Resolver, tracker *resolve.Tracker, def *resolve.Def) error {
	here := collector.Here
	use := collector.Collect

	for _, x := range def.Args {
		shape := tracker.ExtractComponentFactoryShape(x)
		sym := resolver.Symbol(here, shape)
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

	// TODO:
	// if err := tinypkg.Walk(provider, use); err != nil {
	// 	return nil, err
	// }
	return nil
}

func WriteHandlerFunc(w io.Writer,
	here *tinypkg.Package,
	resolver *resolve.Resolver,
	tracker *resolve.Tracker,
	info *web.PathInfo,
	providerModule *resolve.Module,
	runtimeModule *resolve.Module,
	name string,
) error {
	// TODO: typed
	createHandlerFunc, err := providerModule.Type("createHandler")
	if err != nil {
		return fmt.Errorf("in provider module, %w", err)
	}
	createHandlerFunc.Args[0].Name = "getProvider" // todo: remove
	getProviderFunc := createHandlerFunc.Args[0].Node.(*tinypkg.Func)
	getProviderFunc.Name = "getProvider" // todo: remove

	handleResultFunc, err := runtimeModule.Symbol(here, "HandleResult")
	if err != nil {
		return fmt.Errorf("in runtime module, %w", err)
	}
	pathParamFunc, err := runtimeModule.Symbol(here, "PathParam")
	if err != nil {
		return fmt.Errorf("in runtime module, %w", err)
	}

	actionFunc := tinypkg.ToRelativeTypeString(here, info.Def.Symbol)

	provider := &tinypkg.Var{Name: "provider", Node: getProviderFunc.Returns[0].Node}
	if name == "" {
		name = info.Def.Name
	}

	var componentBindings []*tinypkg.Binding
	type pathBinding struct {
		Name string // go's name
		Var  *web.PathVar
	}
	var pathBindings []*pathBinding
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
				pathBindings = append(pathBindings, &pathBinding{Name: x.Name, Var: v})
			}
		default:
			// args = append(args, &tinypkg.Var{Name: x.Name, Node: sym})
		}
	}

	if len(pathBindings) != len(info.VarNames) {
		return fmt.Errorf("invalid path bindings, routing=%v, args=%v (in %s)", info.VarNames, pathBindings, info.Def.Symbol)
	}
	return tinypkg.WriteFunc(w, here, name, createHandlerFunc, func() error {
		fmt.Fprintln(w, "\treturn func(w http.ResponseWriter, req *http.Request) {")
		defer fmt.Fprintln(w, "\t}")

		// <path name> := runtime.PathParam(req, "<path name>")
		if len(pathBindings) > 0 {
			for _, b := range pathBindings {
				// TODO: type check
				fmt.Fprintf(w, "\t\t%s := %s(req, %q)\n", b.Name, pathParamFunc, b.Var.Name)
			}
		}

		// var <component> <type>
		// {
		// 	<component> = <provider>.<method>()
		// }
		if len(componentBindings) > 0 || len(ignored) > 0 {
			if len(componentBindings) == 0 {
				provider.Name = "_"
			}

			fmt.Fprintf(w, "\t\treq, %s, err := %s(req)\n", provider.Name, getProviderFunc.Name)
			fmt.Fprintln(w, "\t\tif err != nil {")
			fmt.Fprintf(w, "\t\t\t%s(w, req, nil, err)\n", handleResultFunc)
			fmt.Fprintln(w, "\t\t\treturn")
			fmt.Fprintln(w, "\t\t}")

			// handling ignored (context.COntext)
			if len(ignored) > 0 {
				for _, x := range ignored {
					if x.Name != "ctx" {
						return fmt.Errorf("unsupported %+v", x)
					}
					fmt.Fprintf(w, "\t\t%s := req.Context()\n", x.Name)
				}
			}

			// handling components
			if len(componentBindings) > 0 {
				indent := "\t\t"
				var returns []*tinypkg.Var
				zeroReturnsDefault := fmt.Sprintf("%s(w, req, nil, err); return", handleResultFunc)
				for _, binding := range componentBindings {
					binding.ZeroReturnsDefault = zeroReturnsDefault
					if err := binding.WriteWithCleanupAndError(w, here, indent, returns); err != nil {
						return err
					}
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
