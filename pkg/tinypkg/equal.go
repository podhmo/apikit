package tinypkg

import "fmt"

func GetLevel(x Node) int {
	if x, ok := x.(*Pointer); ok {
		return x.Lv
	}
	return 0
}

func TypeEqualWithoutLevel(x, y Node) bool {
	if v, ok := x.(*Pointer); ok {
		x = v.V
	}
	if v, ok := y.(*Pointer); ok {
		y = v.V
	}
	return TypeEqual(x, y)
}

func TypeEqual(x, y Node) bool {
	switch x := x.(type) {
	case *Var:
		y, ok := y.(*Var)
		if !ok {
			return false
		}
		return TypeEqual(x.Node, y.Node)
	case *Pointer:
		y, ok := y.(*Pointer)
		if !ok {
			return false
		}
		return x.Lv == y.Lv && TypeEqual(x.V, y.V)
	case *Array:
		y, ok := y.(*Array)
		if !ok {
			return false
		}
		return x.N == y.N && TypeEqual(x.V, y.V)
	case *Slice:
		y, ok := y.(*Slice)
		if !ok {
			return false
		}
		return TypeEqual(x.V, y.V)
	case *Map:
		y, ok := y.(*Map)
		if !ok {
			return false
		}
		return TypeEqual(x.K, y.K) && TypeEqual(x.V, y.V)
	case *Func:
		y, ok := y.(*Func)
		if !ok {
			return false
		}
		if x.Package != y.Package {
			return false
		}
		if x.Name != "" && x.Name != y.Name {
			return false
		}
		if len(x.Args) != len(y.Args) {
			return false
		}
		for i, v := range x.Args {
			if !TypeEqual(v, y.Args[i]) {
				return false
			}
		}
		if len(x.Returns) != len(y.Returns) {
			return false
		}
		for i, v := range x.Returns {
			if !TypeEqual(v, y.Returns[i]) {
				return false
			}
		}
		return true
	case *Interface:
		y, ok := y.(*Interface)
		if !ok {
			return false
		}
		if x.Package != y.Package {
			return false
		}
		if x.Name != "" && x.Name != y.Name {
			return false
		}
		if len(x.Methods) != len(y.Methods) {
			return false
		}
		for i, v := range x.Methods {
			if !TypeEqual(v, y.Methods[i]) {
				return false
			}
		}
		return true
	case *Symbol:
		return x == y
	case *ImportedSymbol:
		y, ok := y.(*ImportedSymbol)
		if !ok {
			return false
		}
		return x.pkg.here == y.pkg.here && TypeEqual(x.sym, y.sym)
	default:
		panic(fmt.Sprintf("unsupported type %T, %T", x, y))
	}
}
