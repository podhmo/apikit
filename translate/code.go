package translate

import (
	"fmt"
	"io"

	"github.com/podhmo/apikit/pkg/tinypkg"
)

type Code struct {
	Name string
	Here *tinypkg.Package

	ImportPackages func() ([]*tinypkg.ImportedPackage, error)
	EmitCode       func(w io.Writer) error
	emitFunc       func(w io.Writer, code *Code) error
}

var ErrNoImports = fmt.Errorf("no imports")

func (c *Code) EmitImports(w io.Writer) error {
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
	return c.emitFunc(w, c)
}
