package tinypkg

import "sync"

type Universe struct {
	packages map[string]*Package
	mu       sync.Mutex
}

func (u *Universe) Lookup(path string) *Package {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.packages[path]
}

func (u *Universe) NewPackage(path, name string) *Package {
	u.mu.Lock()
	defer u.mu.Unlock()
	pkg, ok := u.packages[path]
	if ok {
		return pkg
	}
	pkg = &Package{
		Path:     path,
		Name:     name,
		symbols:  map[string]*Symbol{},
		universe: u,
	}
	u.packages[path] = pkg
	return pkg
}

func NewUniverse() *Universe {
	return &Universe{packages: map[string]*Package{}}
}

type Package struct {
	Path string
	Name string

	universe *Universe
	symbols  map[string]*Symbol
	mu       sync.Mutex
}

func (p *Package) Import(pkg *Package) *ImportedPackage {
	return &ImportedPackage{here: p, pkg: pkg}
}
func (p *Package) ImportAs(pkg *Package, name string) *ImportedPackage {
	return &ImportedPackage{here: p, pkg: pkg, qualifier: name}
}

type ImportedPackage struct {
	here      *Package
	pkg       *Package
	qualifier string
}

func (ip *ImportedPackage) Lookup(sym *Symbol) *ImportedSymbol {
	return &ImportedSymbol{pkg: ip, Name: sym.Name}
}

type ImportedSymbol struct {
	pkg  *ImportedPackage
	Name string
}

func (im *ImportedSymbol) String() string {
	if im.pkg.here == im.pkg.pkg {
		return im.Name
	}
	if im.pkg.qualifier == "" {
		return im.pkg.pkg.Name + "." + im.Name
	}
	return im.pkg.qualifier + "." + im.Name
}

func (p *Package) NewSymbol(name string) *Symbol {
	p.mu.Lock()
	defer p.mu.Unlock()
	sym, ok := p.symbols[name]
	if ok {
		return sym
	}
	sym = &Symbol{Name: name, Package: p}
	p.symbols[name] = sym
	return sym
}

func (p *Package) Lookup(name string) *Symbol {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.symbols[name]
}

type Symbol struct {
	Name    string
	Package *Package
}

func (s *Symbol) String() string {
	return s.Name
}
func (s *Symbol) GoSting() string {
	return s.Package.Name + "." + s.Name
}
