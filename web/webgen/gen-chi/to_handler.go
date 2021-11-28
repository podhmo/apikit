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
	"github.com/podhmo/apikit/web/webgen"
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

			analyzed, err := webgen.Analyze(
				here,
				t.Resolver, t.Tracker,
				pathinfo, extraDefs,
				t.ProviderModule,
			)
			if err != nil {
				return fmt.Errorf("analyze %w", err)
			}

			if len(analyzed.Bindings.Path) != len(pathinfo.VarNames) {
				return fmt.Errorf("invalid path bindings, routing=%v, args=%v (in %s)", pathinfo.VarNames, analyzed.Bindings.Path, pathinfo.Def.Symbol)
			}
			if len(analyzed.Bindings.Data) > 1 {
				return fmt.Errorf("invalid data bindings, support only 1 struct, but found %d (in %s)", len(analyzed.Bindings.Data), pathinfo.Def.Symbol)
			}

			if name == "" {
				name = analyzed.Names.Name
			}
			return WriteHandlerFunc(w, here, t.Resolver,
				analyzed,
				t.RuntimeModule,
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
	analyzed *webgen.Analyzed,
	runtimeModule *resolve.Module,
	name string,
) error {
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

	componentBindings := analyzed.Bindings.Component
	pathBindings := analyzed.Bindings.Path
	queryBindings := analyzed.Bindings.Query
	dataBindings := analyzed.Bindings.Data

	ignored := analyzed.Vars.Ignored
	createHandlerFunc := analyzed.Vars.CreateHandlerFunc

	return tinypkg.WriteFunc(w, here, name, createHandlerFunc, func() error {
		fmt.Fprintln(w, "\treturn func(w http.ResponseWriter, req *http.Request) {")
		defer fmt.Fprintln(w, "\t}")

		// handling path params
		//
		// var pathParams struct { <var 1> string `path:"<var 1>,required"`; ... }
		// runtime.BindPathParams(&pathParams, req, "<var 1>", ...);
		if len(pathBindings) > 0 {
			indent := "\t\t"
			pathParamsName := analyzed.Names.PathParams

			fmt.Fprintf(w, "%svar %s struct {\n", indent, pathParamsName)
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
			indent := "\t\t"
			provider := analyzed.Vars.Provider
			getProviderFunc := analyzed.Vars.GetProviderFunc

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
			queryParamsName := analyzed.Names.QueryParams

			fmt.Fprintf(w, "%svar %s struct {\n", indent, queryParamsName)
			for _, b := range queryBindings {
				fmt.Fprintf(w, "%s\t%s %s `query:\"%s\"`\n", indent, b.Name, b.Sym, b.Name)
			}
			fmt.Fprintf(w, "%s}\n", indent)
			fmt.Fprintf(w, "%sif err := %s(&%s, req); err != nil {\n", indent, bindQueryFunc, queryParamsName)
			fmt.Fprintf(w, "\t%s_ = err // ignored\n", indent)
			fmt.Fprintf(w, "%s}\n", indent)
		}

		// result, err := <action>(....)
		actionName := analyzed.Names.ActionFunc
		actionArgs := analyzed.Names.ActionFuncArgs
		fmt.Fprintf(w, "\t\tresult, err := %s(%s)\n", actionName, strings.Join(actionArgs, ", "))

		// runtime.HandleResult(w, req, result, err)
		fmt.Fprintf(w, "\t\t%s(w, req, result, err)\n", handleResultFunc)
		return nil
	})
}
