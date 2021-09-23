package tinypkg

import (
	"path"
	"strings"
	"sync"
)

type Universe struct {
	packages map[string]*Package
	mu       sync.Mutex
}

func (u *Universe) Lookup(pkgpath string) *Package {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.packages[pkgpath]
}

func (u *Universe) NewPackage(pkgpath, name string) *Package {
	u.mu.Lock()
	defer u.mu.Unlock()
	pkg, ok := u.packages[pkgpath]
	if ok {
		return pkg
	}
	if name == "" {
		parts := strings.Split(pkgpath, "/")
		name = parts[len(parts)-1]
	}

	pkg = &Package{
		Path:     pkgpath,
		Name:     name,
		symbols:  map[string]*Symbol{},
		universe: u,
	}
	u.packages[pkgpath] = pkg
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

func (p *Package) Relative(pkgpath string, name string) *Package {
	fullpath := path.Join(p.Path, pkgpath)
	return p.universe.NewPackage(fullpath, name)
}

func (p *Package) Import(pkg *Package) *ImportedPackage {
	return &ImportedPackage{here: p, pkg: pkg}
}
func (p *Package) ImportAs(pkg *Package, name string) *ImportedPackage {
	return &ImportedPackage{here: p, pkg: pkg, qualifier: name}
}

func (p *Package) NewFunc(name string, args []*Var, returns []*Var) *Func {
	return &Func{
		Name:    name,
		Package: p,
		Args:    args,
		Returns: returns,
	}
}
func (p *Package) NewInterface(name string, fns []*Func) *Interface {
	methods := make([]*Func, len(fns))
	for i, fn := range fns {
		methods[i] = p.NewFunc(fn.Name, fn.Args, fn.Returns)
	}
	return &Interface{
		Name:    name,
		Package: p,
		Methods: methods,
	}
}

type ImportedPackage struct {
	here      *Package
	pkg       *Package
	qualifier string
}

func (ip *ImportedPackage) Lookup(sym *Symbol) *ImportedSymbol {
	return &ImportedSymbol{pkg: ip, sym: sym}
}

type ImportedSymbol struct {
	pkg *ImportedPackage
	sym *Symbol
}

func (im *ImportedSymbol) Qualifier() string {
	qualifier := im.pkg.qualifier
	if qualifier != "" {
		return qualifier
	}
	return im.pkg.pkg.Name
}

func (im *ImportedSymbol) String() string {
	return ToRelativeTypeString(im.pkg.here, im)
}

func (im *ImportedSymbol) onWalk(use func(*Symbol) error) error {
	return use(im.sym)
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
func (s *Symbol) GoString() string {
	if s.Package == nil {
		return s.Name
	}
	return s.Package.Path + "." + s.Name
}
func (s *Symbol) onWalk(use func(*Symbol) error) error {
	return use(s)
}

type Pointer struct {
	Lv int
	V  Node
}

func (c *Pointer) String() string {
	return ToRelativeTypeString(nil, c)
}
func (c *Pointer) onWalk(use func(*Symbol) error) error {
	return c.V.onWalk(use)
}

type Map struct {
	K Node
	V Node
}

func (c *Map) String() string {
	return ToRelativeTypeString(nil, c)
}

func (c *Map) onWalk(use func(*Symbol) error) error {
	if err := c.K.onWalk(use); err != nil {
		return err
	}
	return c.V.onWalk(use)
}

type Slice struct {
	V Node
}

func (c *Slice) String() string {
	return ToRelativeTypeString(nil, c)
}
func (c *Slice) onWalk(use func(*Symbol) error) error {
	return c.V.onWalk(use)
}

type Array struct {
	V Node
	N int
}

func (c *Array) String() string {
	return ToRelativeTypeString(nil, c)
}
func (c *Array) onWalk(use func(*Symbol) error) error {
	return c.V.onWalk(use)
}

type Func struct {
	Name    string
	Recv    string
	Package *Package
	Args    []*Var
	Returns []*Var
}

func (f *Func) onWalk(use func(*Symbol) error) error {
	for _, x := range f.Args {
		if err := x.Node.onWalk(use); err != nil {
			return err
		}
	}
	for _, x := range f.Returns {
		if err := x.Node.onWalk(use); err != nil {
			return err
		}
	}
	return nil
}

func (f *Func) String() string {
	return ToRelativeTypeString(nil, f)
}

type Interface struct {
	Name    string
	Package *Package
	Methods []*Func
}

func (i *Interface) onWalk(use func(*Symbol) error) error {
	for _, x := range i.Methods {
		if err := x.onWalk(use); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interface) String() string {
	return ToRelativeTypeString(nil, i)
}

type Var struct {
	Name string
	Node
}

func (v *Var) String() string {
	return ToRelativeTypeString(nil, v)
}

var (
	_ Node = &Symbol{}
	_ Node = &ImportedSymbol{}
	_ Node = &Pointer{}
	_ Node = &Slice{}
	_ Node = &Array{}
	_ Node = &Map{}
	_ Node = &Func{}
	_ Node = &Interface{}
	_ Node = &Var{}
)
