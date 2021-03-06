package code

import (
	"fmt"
	"go/format"
	"io"

	"github.com/podhmo/apikit/pkg/emitfile"
	"github.com/podhmo/apikit/pkg/tinypkg"
)

var ErrNoImports = fmt.Errorf("no imports")

type Code struct {
	Name   string
	Header string

	Here     *tinypkg.Package
	imported []*tinypkg.ImportedPackage

	ImportPackages func(*tinypkg.ImportCollector) error
	EmitCode       func(w io.Writer, c *Code) error // currently used by Config.EmitCodeFunc

	Priority int
	Config   *Config
	depends  []tinypkg.Node
}

func (c *Code) Import(pkg *tinypkg.Package) *tinypkg.ImportedPackage {
	im := c.Here.Import(pkg)
	c.imported = append(c.imported, im)
	return im
}
func (c *Code) AddDependency(deps ...tinypkg.Node) {
	c.depends = append(c.depends, deps...)
}
func (c *Code) Dependencies() []tinypkg.Node {
	return c.depends
}

// EmitContent emits content only (if you want fullset version, please use c.Config.EmitCodeFunc() )
func (c *Code) EmitContent(w io.Writer) error {
	if err := c.EmitCode(w, c); err != nil {
		return fmt.Errorf("emit content in code %q : %w", c.Name, err)
	}
	return nil
}

// CollectImports is currently used by Config.EmitCodeFunc
func (c *Code) CollectImports(collector *tinypkg.ImportCollector) error {
	if c.Here != collector.Here {
		collector.Add(collector.Here.Import(c.Here))
	}
	return c.collectImportsInner(collector, nil, nil)
}

func (c *Code) collectImportsInner(
	collector *tinypkg.ImportCollector,
	seen map[tinypkg.Node]bool,
	history []*Code,
) error {
	if len(c.imported) > 0 {
		if err := collector.Merge(c.imported); err != nil {
			return err
		}
	}
	if c.depends == nil && c.ImportPackages == nil {
		return nil
	}

	if c.ImportPackages != nil {
		if err := c.ImportPackages(collector); err != nil {
			return fmt.Errorf("in import package : %w", err)
		}
	}
	if c.depends != nil {
		seen = make(map[tinypkg.Node]bool, len(c.depends)+len(history))
		for _, x := range history {
			seen[x] = true
		}
		seen[c] = true
		for _, dep := range c.depends {
			if _, ok := seen[dep]; ok {
				continue
			}
			if child, ok := dep.(*Code); ok {
				if err := child.collectImportsInner(collector, seen, append(history, c)); err != nil {
					return fmt.Errorf("in child code %s: %w", child, err)
				}
			}
			seen[dep] = true
			if err := tinypkg.Walk(dep, collector.Collect); err != nil {
				return fmt.Errorf("in walk : %w", err)
			}
		}
	}
	return nil
}

// String : for pkg/tinypkg.Node
func (c *Code) String() string {
	return tinypkg.ToRelativeTypeString(c.Here, c.Here.NewSymbol(c.Name))
}

// OnWalk : for pkg/tinypkg.Node
func (c *Code) OnWalk(use func(*tinypkg.Symbol) error) error {
	return use(c.Here.NewSymbol(c.Name))
}

// String : for pkg/tinypkg 's internal interface
func (c *Code) RelativeTypeString(here *tinypkg.Package) string {
	return tinypkg.ToRelativeTypeString(here, c.Here.NewSymbol(c.Name))
}

type CodeEmitter struct {
	*Code
}

// EmitImports is currently used by Config.EmitCodeFunc
func (c *CodeEmitter) EmitImports(w io.Writer, imports []*tinypkg.ImportedPackage) error {
	if imports == nil {
		return nil
	}

	io.WriteString(w, "import (\n")
	for _, impkg := range imports {
		fmt.Fprintf(w, "\t%s\n", tinypkg.ToImportPackageString(impkg))
	}
	io.WriteString(w, ")\n")
	return nil
}

// Emit : for pkg/emitfile.Emitter
func (c *CodeEmitter) Emit(w io.Writer) error {
	// emit with header (package declaration, import area)
	return c.Config.EmitCodeFunc(w, c)
}

// Priority : for pkg/emitfile 's internal interface
func (c *CodeEmitter) Priority(priority *int) int {
	return c.Code.Priority
}

// Priority : for pkg/emitfile 's internal interface
func (c *CodeEmitter) FormatBytes(b []byte) ([]byte, error) {
	if c.Config.DisableFormat {
		return b, nil
	}
	return format.Source(b) // TODO: speed-up
}

var _ emitfile.Emitter = &CodeEmitter{}
var _ tinypkg.Node = &Code{}
var _ codeLike = &Code{}
