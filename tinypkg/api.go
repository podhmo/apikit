package tinypkg

var universe = NewUniverse()
var builtins = universe.NewPackage("", "")

func NewPackage(path, name string) *Package {
	return universe.NewPackage(path, name)
}

func NewSymbol(name string) *Symbol {
	return builtins.NewSymbol(name)
}
