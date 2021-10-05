package tinypkg

import "fmt"

func NewImportCollector(here *Package) *ImportCollector {
	return &ImportCollector{
		Here: here,
		seen: map[*Package]bool{},
	}
}

type ImportCollector struct {
	Here    *Package
	Imports []*ImportedPackage
	seen    map[*Package]bool
}

func (c *ImportCollector) Add(im *ImportedPackage) error {
	if _, ok := c.seen[im.pkg]; ok {
		return nil
	}
	if im.here != c.Here {
		return fmt.Errorf("imported package is mismatched (here is %q, but imported is %q)", c.Here.Path, im.here.Path)
	}
	c.seen[im.pkg] = true
	c.Imports = append(c.Imports, im)
	return nil
}

func (c *ImportCollector) Merge(imports []*ImportedPackage) error {
	for _, im := range imports {
		if _, ok := c.seen[im.pkg]; ok {
			continue
		}
		if im.here != c.Here {
			return fmt.Errorf("imported package is mismatched (here is %q, but imported is %q)", c.Here.Path, im.here.Path)
		}
		c.seen[im.pkg] = true
		c.Imports = append(c.Imports, im)
	}
	return nil
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
