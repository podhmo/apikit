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
	Depends  []tinypkg.Node
}

func (c *Code) Import(pkg *tinypkg.Package) *tinypkg.ImportedPackage {
	im := c.Here.Import(pkg)
	c.imported = append(c.imported, im)
	return im
}
func (c *Code) AddDependency(dep tinypkg.Node) {
	c.Depends = append(c.Depends, dep)
}

// CollectImports is currently used by Config.EmitCodeFunc
func (c *Code) CollectImports(collector *tinypkg.ImportCollector) error {
	if c.Here != collector.Here {
		collector.Add(collector.Here.Import(c.Here))
	}
	if len(c.imported) > 0 {
		if err := collector.Merge(c.imported); err != nil {
			return err
		}
	}
	if c.Depends == nil && c.ImportPackages == nil {
		return nil
	}

	if c.ImportPackages != nil {
		if err := c.ImportPackages(collector); err != nil {
			return fmt.Errorf("in import package : %w", err)
		}
	}
	if c.Depends != nil { // TODO: cache
		seen := make(map[tinypkg.Node]struct{}, len(c.Depends))
		for _, dep := range c.Depends {
			if _, ok := seen[dep]; ok {
				continue
			}
			seen[dep] = struct{}{}
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
