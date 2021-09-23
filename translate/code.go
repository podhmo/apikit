package translate

import (
	"fmt"
	"go/format"
	"io"

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
}

func (c *Code) Priority() int {
	return c.priority
}

func (c *Code) FormatBytes(b []byte) ([]byte, error) {
	return format.Source(b) // TODO: speed-up
}

func (c *Code) EmitImports(w io.Writer) error {
	if c.ImportPackages == nil {
		return nil
	}

	impkgs, err := c.ImportPackages()
	if err != nil {
		return err
	}
	if len(impkgs) == 0 {
		return ErrNoImports
	}

	io.WriteString(w, "import (\n")
	for _, impkg := range impkgs {
		fmt.Fprintf(w, "\t%s\n", tinypkg.ToImportPackageString(impkg))
	}
	io.WriteString(w, ")\n")
	return nil
}

// for pkg/emitfile.Emitter
func (c *Code) Emit(w io.Writer) error {
	return c.Config.EmitCodeFunc(w, c)
}
