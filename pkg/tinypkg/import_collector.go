package tinypkg

type ImportCollector struct {
	Here    *Package
	Imports []*ImportedPackage
	seen    map[*Package]bool
}

func (c *ImportCollector) Collect(sym *Symbol) error {
	seen := c.seen
	here := c.Here
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
	c.Imports = append(c.Imports, here.Import(sym.Package))
	return nil
}
