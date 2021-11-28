package genchi

import (
	"fmt"
	"io"
	"strconv"
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

	if t.Config.Verbose {
		t.Config.Log.Printf("\t+ translate %s.%s -> handler %s.%s", def.Package.Path, def.Symbol, here.Path, name)
	}

	extraDeps := web.GetMetaData(node.Node).ExtraDependencies
	extraDefs := make([]*resolve.Def, len(extraDeps))
	for i, fn := range extraDeps {
		extraDef := t.Resolver.Def(fn)
		t.Tracker.Track(extraDef)
		extraDefs[i] = extraDef
	}

	c := &code.Code{
		Name: name,
		Here: here,
		// priority: code.PrioritySecond,
		Config: t.Config,
		ImportPackages: func(collector *tinypkg.ImportCollector) error {
			// todo: support provider *tinypkg.Var
			if err := collectImportsForHandler(collector, t.Resolver, t.Tracker, def); err != nil {
				return err
			}
			if len(extraDefs) > 0 {
				for _, extraDef := range extraDefs {
					if err := collectImportsForHandler(collector, t.Resolver, t.Tracker, extraDef); err != nil {
						return err
					}
				}
			}
			return nil
		},
		EmitCode: func(w io.Writer, c *code.Code) error {
			pathinfo, err := web.ExtractPathInfo(node.Node.VariableNames, def)
			if err != nil {
				return err
			}
			c.AddDependency(t.ProviderModule)
			c.AddDependency(t.RuntimeModule)
			return WriteHandlerFunc(w, here,
				t.Resolver, t.Tracker,
				pathinfo, extraDefs,
				t.ProviderModule, t.RuntimeModule,
				name,
			)
		},
	}
	return &code.CodeEmitter{Code: c}
}

func collectImportsForHandler(collector *tinypkg.ImportCollector, resolver *resolve.Resolver, tracker *resolve.Tracker, def *resolve.Def) error {
	here := collector.Here
	use := collector.Collect

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

func WriteHandlerFunc(w io.Writer,
	here *tinypkg.Package,
	resolver *resolve.Resolver,
	tracker *resolve.Tracker,
	info *web.PathInfo,
	extraDefs []*resolve.Def,
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
	bindPathParamsFunc, err := runtimeModule.Symbol(here, "BindPathParams")
	if err != nil {
		return fmt.Errorf("in runtime module, %w", err)
	}
	bindQueryFunc, err := runtimeModule.Symbol(here, "BindQuery")
	if err != nil {
		return fmt.Errorf("in runtime module, %w", err)
	}
	bindBodyFunc, err := runtimeModule.Symbol(here, "BindBody")
	if err != nil {
		return fmt.Errorf("in runtime module, %w", err)
	}
	validateStructFunc, err := runtimeModule.Symbol(here, "ValidateStruct")
	if err != nil {
		return fmt.Errorf("in runtime module, %w", err)
	}

	actionFunc := tinypkg.ToRelativeTypeString(here, info.Def.Symbol)

	provider := &tinypkg.Var{Name: "provider", Node: getProviderFunc.Returns[0].Node}
	if name == "" {
		name = info.Def.Name
	}

	var componentBindings tinypkg.BindingList

	type pathBinding struct {
		Name string // go's name
		Var  *web.PathVar
		Sym  tinypkg.Node
	}
	var pathBindings []*pathBinding
	type queryBinding struct {
		Name string
		Sym  tinypkg.Node
	}
	var queryBindings []*queryBinding
	var dataBindings []resolve.Item

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
				return err
			}

			rt := x.Shape.GetReflectType() // not shape.GetRefectType()
			methodName := tracker.ExtractMethodName(rt, x.Name)
			binding.ProviderAlias = fmt.Sprintf("%s.%s", provider.Name, methodName)

			componentBindings = append(componentBindings, binding)
			argNames = append(argNames, x.Name)
		case resolve.KindPrimitive: // handle pathParams
			if v, ok := info.Vars[x.Name]; ok {
				pathBindings = append(pathBindings, &pathBinding{Name: x.Name, Var: v, Sym: resolver.Symbol(here, v.Shape)})
			}
			argNames = append(argNames, "pathParams."+x.Name)
		case resolve.KindPrimitivePointer: // handle query string
			queryBindings = append(queryBindings, &queryBinding{Name: x.Name, Sym: resolver.Symbol(here, x.Shape)})
			argNames = append(argNames, "queryParams."+x.Name)
		case resolve.KindPrimitiveSlicePointer: // handle query string
			queryBindings = append(queryBindings, &queryBinding{Name: x.Name, Sym: resolver.Symbol(here, x.Shape)})
			argNames = append(argNames, "queryParams."+x.Name)
		case resolve.KindData: // handle request.Body
			dataBindings = append(dataBindings, x)
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
						return err
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
				return err
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
						return err
					}

					rt := subShape.GetReflectType() // not shape.GetRefectType()
					methodName := tracker.ExtractMethodName(rt, name)
					binding.ProviderAlias = fmt.Sprintf("%s.%s", provider.Name, methodName)

					componentBindings = append(componentBindings, binding)
				}
			}
		}
	}

	if len(pathBindings) != len(info.VarNames) {
		return fmt.Errorf("invalid path bindings, routing=%v, args=%v (in %s)", info.VarNames, pathBindings, info.Def.Symbol)
	}
	if len(dataBindings) > 1 {
		return fmt.Errorf("invalid data bindings, support only 1 struct, but found %d (in %s)", len(dataBindings), info.Def.Symbol)
	}

	return tinypkg.WriteFunc(w, here, name, createHandlerFunc, func() error {
		fmt.Fprintln(w, "\treturn func(w http.ResponseWriter, req *http.Request) {")
		defer fmt.Fprintln(w, "\t}")

		// handling path params
		//
		// var pathParams struct { <var 1> string `path:"<var 1>,required"`; ... }
		// runtime.BindPathParams(&pathParams, req, "<var 1>", ...);
		if len(pathBindings) > 0 {
			indent := "\t\t"
			fmt.Fprintf(w, "%svar pathParams struct {\n", indent)
			varNames := make([]string, len(pathBindings))
			for i, b := range pathBindings {
				fmt.Fprintf(w, "%s\t%s %s `query:\"%s,required\"`\n", indent, b.Name, b.Sym, b.Var.Name)
				varNames[i] = strconv.Quote(b.Var.Name)
			}
			fmt.Fprintf(w, "%s}\n", indent)
			fmt.Fprintf(w, "%sif err := %s(&pathParams, req, %s); err != nil {\n", indent, bindPathParamsFunc, strings.Join(varNames, ", "))
			fmt.Fprintf(w, "%s\tw.WriteHeader(404)\n", indent)
			fmt.Fprintf(w, "\t%s%s(w, req, nil, err); return\n", indent, handleResultFunc)
			fmt.Fprintf(w, "%s}\n", indent)
		}

		// var <component> <type>
		// {
		// 	<component> = <provider>.<method>()
		// }
		if len(componentBindings) > 0 || len(ignored) > 0 {
			if len(componentBindings)-len(extraDefs) == 0 {
				provider.Name = "_"
			}

			indent := "\t\t"
			fmt.Fprintf(w, "%sreq, %s, err := %s(req)\n", indent, provider.Name, getProviderFunc.Name)
			fmt.Fprintf(w, "%sif err != nil {\n", indent)
			fmt.Fprintf(w, "%s\t%s(w, req, nil, err); return\n", indent, handleResultFunc)
			fmt.Fprintf(w, "%s}\n", indent)

			// handling ignored (context.Context, *http.Request)
			if len(ignored) > 0 {
				for _, x := range ignored {
					if x.Name == "ctx" {
						fmt.Fprintf(w, "\t\tvar %s context.Context = req.Context()\n", x.Name)
					}
				}
			}

			// handling components
			if len(componentBindings) > 0 {
				indent := "\t\t"
				var returns []*tinypkg.Var
				zeroReturnsDefault := fmt.Sprintf("%s(w, req, nil, err); return", handleResultFunc)
				sorted, err := componentBindings.TopologicalSorted(ignored...)
				if err != nil {
					return fmt.Errorf("failed component binding (toposort): %w", err)
				}
				for _, binding := range sorted {
					binding.ZeroReturnsDefault = zeroReturnsDefault
					if err := binding.WriteWithCleanupAndError(w, here, indent, returns); err != nil {
						return err
					}
				}
			}
		}

		// handling request body
		// var data <struct>
		// runtime.Bind(data, req.Body)
		// runtime.ValidateStruct(data)
		if len(dataBindings) > 0 {
			indent := "\t\t"
			x := dataBindings[0]
			fmt.Fprintf(w, "%svar %s %s\n", indent, x.Name, resolver.Symbol(here, x.Shape)) // todo: depenency?

			fmt.Fprintf(w, "%sif err := %s(&%s, req.Body); err != nil {\n", indent, bindBodyFunc, x.Name)
			fmt.Fprintf(w, "\t%sw.WriteHeader(400)\n", indent)
			fmt.Fprintf(w, "\t%s%s(w, req, nil, err); return\n", indent, handleResultFunc)
			fmt.Fprintf(w, "%s}\n", indent)

			fmt.Fprintf(w, "%sif err := %s(&%s); err != nil {\n", indent, validateStructFunc, x.Name)
			fmt.Fprintf(w, "\t%sw.WriteHeader(422)\n", indent)
			fmt.Fprintf(w, "\t%s%s(w, req, nil, err); return\n", indent, handleResultFunc)
			fmt.Fprintf(w, "%s}\n", indent)
		}

		// handling query params
		//
		// var queryParams struct { <var 1> string `query:"<var 1>,required"`; ... }
		// runtime.BindQuery(&queryParams, req);
		if len(queryBindings) > 0 {
			indent := "\t\t"
			fmt.Fprintf(w, "%svar queryParams struct {\n", indent)
			for _, b := range queryBindings {
				fmt.Fprintf(w, "%s\t%s %s `query:\"%s\"`\n", indent, b.Name, b.Sym, b.Name)
			}
			fmt.Fprintf(w, "%s}\n", indent)
			fmt.Fprintf(w, "%sif err := %s(&queryParams, req); err != nil {\n", indent, bindQueryFunc)
			fmt.Fprintf(w, "\t%s_ = err // ignored\n", indent)
			fmt.Fprintf(w, "%s}\n", indent)
		}

		// result, err := <action>(....)
		fmt.Fprintf(w, "\t\tresult, err := %s(%s)\n", actionFunc, strings.Join(argNames, ", "))

		// runtime.HandleResult(w, req, result, err)
		fmt.Fprintf(w, "\t\t%s(w, req, result, err)\n", handleResultFunc)
		return nil
	})
}
