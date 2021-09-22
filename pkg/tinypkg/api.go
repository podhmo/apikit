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

func ColletImprts(x Node, here *Package) ([]*ImportedPackage, error) {
	c := &ImportCollector{
		Here: here,
		seen: map[*Package]bool{},
	}
	if err := Walk(x, c.Collect); err != nil {
		return nil, err
	}
	return c.Imports, nil
}

type Node interface {
	fmt.Stringer
	onWalk(use func(*Symbol) error) error
}
