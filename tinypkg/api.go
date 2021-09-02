package tinypkg

var universe = NewUniverse()
var builtins = universe.NewPackage("", "")

func NewPackage(path, name string) *Package {
	return universe.NewPackage(path, name)
}

func NewSymbol(name string) *Symbol {
	return builtins.NewSymbol(name)
}

func Walk(x Symboler, use func(*Symbol) error) error {
	if v, ok := x.(walkerNode); ok {
		if err := v.onWalk(use); err != nil {
			return err
		}
		return nil
	}
	return use(x.Symbol())
}
