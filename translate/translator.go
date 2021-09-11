package translate

import (
	"io"

	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/tinypkg"
)

type Translator struct {
	Tracker  *Tracker
	Resolver *resolve.Resolver
}

func NewTranslator(resolver *resolve.Resolver, fns []interface{}) *Translator {
	tracker := NewTracker()
	for _, fn := range fns {
		def := resolver.Resolve(fn)
		tracker.Track(def)
	}
	return &Translator{
		Tracker:  tracker,
		Resolver: resolver,
	}
}

func (t *Translator) TranslateInterface(here *tinypkg.Package, name string) *Code {
	return &Code{
		Name: name,
		ImportPackages: func() ([]*tinypkg.ImportedPackage, error) {
			return collectImports(here, t.Tracker)
		},
		EmitCode: func(w io.Writer) error {
			writeInterface(w, here, t.Tracker, name)
			return nil
		},
	}
}

type Code struct {
	Name           string
	ImportPackages func() ([]*tinypkg.ImportedPackage, error)
	EmitCode       func(w io.Writer) error
}

func (c *Code) EmitImports(w io.Writer) error {
	impkgs, err := c.ImportPackages()
	if err != nil {
		return err
	}
	if len(impkgs) == 0 {
		return nil
	}

	io.WriteString(w, "import (\n")
	for _, impkg := range impkgs {
		io.WriteString(w, "\t")
		impkg.Emit(w)
	}
	io.WriteString(w, ")\n")
	return nil
}
