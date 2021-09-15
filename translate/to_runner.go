package translate

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

// TODO: omit provider arguments

func (t *Translator) TranslateToRunner(here *tinypkg.Package, def *resolve.Def, name string, provider *tinypkg.Var) *Code {
	if name == "" {
		name = def.Name
	}
	if provider == nil {
		provider = t.providerVar
	}

	t.Tracker.Track(def)
	return &Code{
		Name:     name,
		Here:     here,
		emitFunc: t.EmitFunc,
		ImportPackages: func() ([]*tinypkg.ImportedPackage, error) {
			return collectImportsForRunner(here, t.Resolver, def, provider)
		},
		EmitCode: func(w io.Writer) error {
			return writeRunner(w, here, t.Resolver, t.Tracker, def, provider, name)
		},
	}
}

func collectImportsForRunner(here *tinypkg.Package, resolver *resolve.Resolver, def *resolve.Def, provider *tinypkg.Var) ([]*tinypkg.ImportedPackage, error) {
	imports := make([]*tinypkg.ImportedPackage, 0, len(def.Args)+len(def.Returns))
	seen := map[*tinypkg.Package]bool{}
	use := func(sym *tinypkg.Symbol) error {
		if sym.Package.Path == "" {
			return nil // bultins type (e.g. string, bool, ...)
		}
		if _, ok := seen[sym.Package]; ok {
			return nil
		}
		seen[sym.Package] = true
		if here == sym.Package {
			return nil
		}
		imports = append(imports, here.Import(sym.Package))
		return nil
	}

	for _, x := range def.Args {
		sym := resolver.Symbol(here, x.Shape)
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
	return imports, nil
}

func writeRunner(w io.Writer, here *tinypkg.Package, resolver *resolve.Resolver, tracker *Tracker, def *resolve.Def, provider *tinypkg.Var, name string) error {
	var components []resolve.Item
	var ignored []*tinypkg.Var
	argNames := make([]string, 0, len(def.Args))
	args := make([]*tinypkg.Var, 0, len(def.Args)+1)
	{
		sym := provider.Node.(*tinypkg.Symbol)
		args = append(args, &tinypkg.Var{Name: provider.Name, Node: here.Import(sym.Package).Lookup(sym)})
	}
	for _, x := range def.Args {
		argNames = append(argNames, x.Name)

		if x.Kind == resolve.KindComponent {
			components = append(components, x)
			continue
		}

		sym := resolver.Symbol(here, x.Shape)
		if x.Kind == resolve.KindIgnored { // e.g. context.Context
			ignored = append(ignored, &tinypkg.Var{Name: x.Name, Node: sym})
		} else {
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

	return tinypkg.WriteFunc(w, name, &tinypkg.Func{Args: args, Returns: returns},
		func() error {
			// var <component> <type>
			// {
			//   <component> = <provider>.<method>()
			// }
			if len(components) > 0 {
				isConsumeFuncReturnsError := len(returns) > 0 && def.Shape.GetReflectType().Out(len(returns)-1) == reflectErrorTypeValue

				for _, x := range components {
					shape := x.Shape
					for _, need := range tracker.seen[x.Shape.GetReflectType()] {
						if need.Name == x.Name {
							if need.overrideDef != nil {
								shape = need.overrideDef.Shape
							}
							break
						}
					}

					sym := resolver.Symbol(here, shape)
					methodName := x.Shape.GetReflectType().Name()

					// var x <X>
					if f, ok := sym.(*tinypkg.Func); ok {
						fmt.Fprintf(w, "\tvar %s %s\n", x.Name, tinypkg.ToRelativeTypeString(here, f.Returns[0]))
					} else {
						fmt.Fprintf(w, "\tvar %s %s\n", x.Name, tinypkg.ToRelativeTypeString(here, sym))
					}

					// {
					fmt.Fprintln(w, "\t{")

					hasError := false
					hasCleanup := false
					switch provided := sym.(type) {
					case *tinypkg.Func:
						switch len(provided.Returns) {
						case 1: // x := provide()
							fmt.Fprintf(w, "\t\t%s = %s.%s()\n", x.Name, provider.Name, methodName)
						case 2: // x, err := provide()
							if provided.Returns[1].Node.String() == "error" {
								hasError = true
								fmt.Fprintln(w, "\t\tvar err error")
								fmt.Fprintf(w, "\t\t%s, err = %s.%s()\n", x.Name, provider.Name, methodName)
							} else {
								hasCleanup = true
								fmt.Fprintf(w, "\t\tvar cleanup %s\n", tinypkg.ToRelativeTypeString(here, provided.Returns[1]))
								fmt.Fprintf(w, "\t\t%s, cleanup = %s.%s()\n", x.Name, provider.Name, methodName)
								if _, ok := provided.Returns[1].Node.(*tinypkg.Func); !ok {
									return fmt.Errorf("unsupported provide function, only support func(...)(<T>, error) or func(...)(<T>, func()). got=%s", provided)
								}
							}
						case 3: // x, cleanup, err := provide()
							hasError = true
							hasCleanup = true
							fmt.Fprintf(w, "\t\tvar cleanup %s\n", tinypkg.ToRelativeTypeString(here, provided.Returns[1]))
							fmt.Fprintln(w, "\t\tvar err error")
							fmt.Fprintf(w, "\t\t%s, cleanup, err = %s.%s()\n", x.Name, provider.Name, methodName)
							if _, ok := provided.Returns[1].Node.(*tinypkg.Func); !ok {
								return fmt.Errorf("unsupported provide function, only support func(...)(<T>, func(), error). got=%s", provided)
							}
						default:
							return fmt.Errorf("unexpected provider function for %s, %+v", x.Name, shape)
						}
					default:
						fmt.Fprintf(w, "\t\t%s = %s.%s()\n", x.Name, provider.Name, methodName)
					}

					if hasCleanup {
						fmt.Fprintln(w, "\t\tif cleanup != nil {")
						fmt.Fprintln(w, "\t\t\tdefer cleanup()") // TODO: support Close() error
						fmt.Fprintln(w, "\t\t}")
					}
					if hasError {
						fmt.Fprintf(w, "\t\tif err != nil {\n")
						switch len(returns) {
						case 0:
							fmt.Fprintln(w, "\t\t\treturn")
						case 1, 2, 3:
							// return <>
							// return error
							// return <>, error
							// return <>, func()
							// return <>, error, func()
							values := []string{"nil", "nil", "nil"}
							if isConsumeFuncReturnsError {
								values[len(returns)-1] = "err"
							}
							fmt.Fprintf(w, "\t\t\treturn %s\n", strings.Join(values[:len(returns)], ", "))
						default:
							return fmt.Errorf("unsupported consume function: %s", def.Symbol)
						}
						fmt.Fprintln(w, "\t\t}")
					}
					// }
					fmt.Fprintln(w, "\t}")
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

var reflectErrorTypeValue = reflect.TypeOf(func() error { return nil }).Out(0)
