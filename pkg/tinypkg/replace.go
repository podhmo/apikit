package tinypkg

import "fmt"

func Replace(node Node, use func(*Symbol) Node) Node {
	return replace(node, use)
}

func replace(node Node, use func(*Symbol) Node) Node {
	switch x := node.(type) {
	case *Var:
		return &Var{Name: x.Name, Node: replace(x.Node, use)}
	case *Pointer:
		return &Pointer{Lv: x.Lv, V: replace(x.V, use)}
	case *Array:
		return &Array{N: x.N, V: replace(x.V, use)}
	case *Slice:
		return &Slice{V: replace(x.V, use)}
	case *Map:
		k := replace(x.K, use)
		v := replace(x.V, use)
		return &Map{K: k, V: v}
	case *Func:
		args := make([]*Var, len(x.Args))
		for i, x := range x.Args {
			args[i] = &Var{Name: x.Name, Node: replace(x.Node, use)}
		}
		returns := make([]*Var, len(x.Returns))
		for i, x := range x.Returns {
			returns[i] = &Var{Name: x.Name, Node: replace(x.Node, use)}
		}
		return &Func{Name: x.Name, Package: x.Package, Args: args, Returns: returns, Recv: x.Recv}
	case *Interface:
		methods := make([]*Func, len(x.Methods))
		for i, method := range x.Methods {
			methods[i] = replace(method, use).(*Func) // xxx
		}
		return &Interface{Name: x.Name, Package: x.Package, Methods: methods}
	case *Symbol:
		return use(x)
	case *ImportedSymbol:
		replaced := replace(x.sym, use)
		if sym, ok := replaced.(*Symbol); ok {
			return x.pkg.here.Import(sym.Package).Lookup(sym)
		}
		return replaced
	default:
		panic(fmt.Sprintf("unsupported type %T", node))
	}
}
