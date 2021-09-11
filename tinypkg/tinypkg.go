package tinypkg

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

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
	if name == "" {
		parts := strings.Split(path, "/")
		name = parts[len(parts)-1]
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
	return &ImportedSymbol{pkg: ip, sym: sym}
}

// TODO: omit
func (ip *ImportedPackage) Emit(w io.Writer) error {
	if ip.qualifier != "" {
		io.WriteString(w, ip.qualifier)
		io.WriteString(w, " ")
	}
	fmt.Fprintf(w, "%q\n", ip.pkg.Path)
	return nil
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

func (im *ImportedSymbol) Symbol() *Symbol {
	return im.sym
}

func (im *ImportedSymbol) String() string {
	return ToRelativeTypeString(im.pkg.here, im)
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

func (s *Symbol) Symbol() *Symbol {
	return s
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

type Pointer struct {
	Lv int
	V  Symboler
}

func (c *Pointer) String() string {
	return ToRelativeTypeString(nil, c)
}
func (c *Pointer) Symbol() *Symbol {
	return c.V.Symbol()
}
func (c *Pointer) onWalk(use func(*Symbol) error) error {
	if v, ok := c.V.(walkerNode); ok {
		return v.onWalk(use)
	}
	return use(c.Symbol())
}

type Map struct {
	K Symboler
	V Symboler
}

func (c *Map) String() string {
	return ToRelativeTypeString(nil, c)
}
func (c *Map) Symbol() *Symbol {
	k := c.K.Symbol()
	if k != nil {
		return k // TODO: return K and V
	}
	return c.V.Symbol()
}
func (c *Map) onWalk(use func(*Symbol) error) error {
	if v, ok := c.K.(walkerNode); ok {
		if err := v.onWalk(use); err != nil {
			return err
		}
	} else {
		if err := use(c.K.Symbol()); err != nil {
			return err
		}
	}
	if v, ok := c.V.(walkerNode); ok {
		return v.onWalk(use)
	}
	return use(c.V.Symbol())
}

type Slice struct {
	V Symboler
}

func (c *Slice) Symbol() *Symbol {
	return c.V.Symbol()
}
func (c *Slice) String() string {
	return ToRelativeTypeString(nil, c)
}
func (c *Slice) onWalk(use func(*Symbol) error) error {
	if v, ok := c.V.(walkerNode); ok {
		return v.onWalk(use)
	}
	return use(c.Symbol())
}

type Array struct {
	V Symboler
	N int
}

func (c *Array) String() string {
	return ToRelativeTypeString(nil, c)
}
func (c *Array) Symbol() *Symbol {
	return c.V.Symbol()
}
func (c *Array) onWalk(use func(*Symbol) error) error {
	if v, ok := c.V.(walkerNode); ok {
		return v.onWalk(use)
	}
	return use(c.Symbol())
}

type Func struct {
	Name    string
	Params  []*Var
	Returns []*Var
}

func (f *Func) Symbol() *Symbol {
	return nil // TODO: this is broken.
}
func (f *Func) onWalk(use func(*Symbol) error) error {
	for _, x := range f.Params {
		if v, ok := x.Symboler.(walkerNode); ok {
			if err := v.onWalk(use); err != nil {
				return err
			}
			continue
		}
		if err := use(x.Symbol()); err != nil {
			return err
		}
	}
	for _, x := range f.Returns {
		if v, ok := x.Symboler.(walkerNode); ok {
			if err := v.onWalk(use); err != nil {
				return err
			}
			continue
		}
		if err := use(x.Symbol()); err != nil {
			return err
		}
	}
	return nil
}

func (f *Func) String() string {
	return ToRelativeTypeString(nil, f)
}

type Var struct {
	Name string
	Symboler
}

func (v *Var) String() string {
	return ToRelativeTypeString(nil, v)
}
