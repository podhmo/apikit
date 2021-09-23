package translate

import (
	"fmt"
	"go/format"
	"io"

	"github.com/podhmo/apikit/pkg/emitfile"
	"github.com/podhmo/apikit/pkg/tinypkg"
)

var ErrNoImports = fmt.Errorf("no imports")

type Code struct {
	Name string
	Here *tinypkg.Package

	ImportPackages func() ([]*tinypkg.ImportedPackage, error)
	EmitCode       func(w io.Writer) error

	priority int
	Config   *Config
	Depends  []tinypkg.Node
}

func (c *Code) Priority() int {
	return c.priority
}

func (c *Code) FormatBytes(b []byte) ([]byte, error) {
	if c.Config.DisableFormat {
		return b, nil
	}
	return format.Source(b) // TODO: speed-up
}

func (c *Code) EmitImports(w io.Writer) error {
	if c.Depends == nil && c.ImportPackages == nil {
		return nil
	}

	var imports []*tinypkg.ImportedPackage
	if c.ImportPackages != nil {
		impkgs, err := c.ImportPackages()
		if err != nil {
			return err
		}
		if len(impkgs) == 0 {
			return ErrNoImports
		}
		imports = append(imports, impkgs...)
	}
	if c.Depends != nil {
		collector := tinypkg.NewImportCollector(c.Here)
		if err := collector.Merge(imports); err != nil {
			return fmt.Errorf("emit import in code %q : %w", c.Name, err)
		}
		for _, dep := range c.Depends {
			if err := tinypkg.Walk(dep, collector.Collect); err != nil {
				return fmt.Errorf("emit import in code %q, in walk : %w", c.Name, err)
			}
		}
		imports = collector.Imports
	}

	io.WriteString(w, "import (\n")
	for _, impkg := range imports {
		fmt.Fprintf(w, "\t%s\n", tinypkg.ToImportPackageString(impkg))
	}
	io.WriteString(w, ")\n")
	return nil
}

// Emit for pkg/emitfile.Emitter
func (c *Code) Emit(w io.Writer) error {
	return c.Config.EmitCodeFunc(w, c)
}

// String for pkg/tinypkg.Node
func (c *Code) String() string {
	return tinypkg.ToRelativeTypeString(c.Here, c.Here.NewSymbol(c.Name))
}

// OnWalk for pkg/tinypkg.Node
func (c *Code) OnWalk(use func(*tinypkg.Symbol) error) error {
	return use(c.Here.NewSymbol(c.Name))
}

// String for pkg/tinypkg 's internal interface
func (c *Code) RelativeTypeString(here *tinypkg.Package) string {
	return tinypkg.ToRelativeTypeString(here, c.Here.NewSymbol(c.Name))
}

var _ emitfile.Emitter = &Code{}
var _ tinypkg.Node = &Code{}
