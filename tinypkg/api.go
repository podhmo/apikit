package tinypkg

import "fmt"

var universe = NewUniverse()
var builtins = universe.NewPackage("", "")

func NewPackage(path, name string) *Package {
	return universe.NewPackage(path, name)
}

func NewSymbol(name string) *Symbol {
	return builtins.NewSymbol(name)
}

func Walk(x Node, use func(*Symbol) error) error {
	return x.onWalk(use)
}

type Node interface {
	fmt.Stringer
	onWalk(use func(*Symbol) error) error
}
