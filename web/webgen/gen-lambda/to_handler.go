package genlambda

import (
	"fmt"
	"io"
	"strings"

	"github.com/podhmo/apikit/code"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

func (t *Translator) TranslateToHandler(here *tinypkg.Package, actionFunc interface{}, name string) *code.CodeEmitter {
	def := t.Resolver.Def(actionFunc)
	if name == "" {
		name = def.Name
	}
	t.Tracker.Track(def)

	if t.Config.Verbose {
		t.Config.Log.Printf("\t+ translate %s.%s -> handler %s.%s", def.Package.Path, def.Symbol, here.Path, name)
	}

	// extraDeps := web.GetMetaData(node.Node).ExtraDependencies
	// extraDefs := make([]*resolve.Def, len(extraDeps))
	// for i, fn := range extraDeps {
	// 	extraDef := t.Resolver.Def(fn)
	// 	t.Tracker.Track(extraDef)
	// 	extraDefs[i] = extraDef
	// }

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
			// if len(extraDefs) > 0 {
			// 	for _, extraDef := range extraDefs {
			// 		if err := collectImportsForHandler(collector, t.Resolver, t.Tracker, extraDef); err != nil {
			// 			return err
			// 		}
			// 	}
			// }
			return nil
		},
		EmitCode: func(w io.Writer, c *code.Code) error {
			c.AddDependency(t.ProviderModule)
			c.AddDependency(t.RuntimeModule)
			return WriteHandlerFunc(w, here,
				t.Resolver, t.Tracker,
				def,
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
	// info *web.PathInfo,
	// extraDefs []*resolve.Def,
	def *resolve.Def,
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

	actionFunc := tinypkg.ToRelativeTypeString(here, def.Symbol)
	var argNames []string // TODO: fix

	fmt.Fprintln(w, "// see: https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html")
	fmt.Fprintln(w, "")

	return tinypkg.WriteFunc(w, here, name, createHandlerFunc, func() error {
		fmt.Fprintf(w, "return func (ctx context.Context, event Event) (interface{}, error) {\n")
		// result, err := <action>(....)
		fmt.Fprintf(w, "\t\tresult, err := %s(%s)\n", actionFunc, strings.Join(argNames, ", "))
		fmt.Fprintf(w, "\t\tif err != nil {\n")
		fmt.Fprintf(w, "\t\t\treturn nil, err")
		fmt.Fprintf(w, "\t\t}\n")
		fmt.Fprintf(w, "\t\treturn result, nil")
		defer fmt.Fprintln(w, "\t}")
		return nil
	})
}
