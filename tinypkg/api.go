package tinypkg

var universe = NewUniverse()

func NewPackage(path, name string) *Package {
	return universe.NewPackage(path, name)
}
