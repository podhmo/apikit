package translate

import (
	"fmt"
	"io"

	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/tinypkg"
)

type Emitter interface {
	Emit(w io.Writer, code *Code) error
}
type EmitFunc func(w io.Writer, code *Code) error

func (f EmitFunc) Emit(w io.Writer, code *Code) error {
	return f(w, code)
}

const Header = `// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

`

func defaultEmitFunc(w io.Writer, code *Code) error {
	fmt.Fprintln(w, Header)
	fmt.Fprintf(w, "package %s\n\n", code.Here.Name)
	if err := code.EmitImports(w); err != nil {
		if err != ErrNoImports {
			return err
		}
	} else {
		io.WriteString(w, "\n")
	}
	return code.EmitCode(w)
}

type Translator struct {
	Tracker  *Tracker
	Resolver *resolve.Resolver
	EmitFunc EmitFunc
}

func NewTranslator(resolver *resolve.Resolver, fns ...interface{}) *Translator {
	tracker := NewTracker()
	for _, fn := range fns {
		def := resolver.Resolve(fn)
		tracker.Track(def)
	}
	return &Translator{
		Tracker:  tracker,
		Resolver: resolver,
		EmitFunc: defaultEmitFunc,
	}
}

func (t *Translator) TranslateInterface(here *tinypkg.Package, name string) *Code {
	return &Code{
		Name:     name,
		Here:     here,
		EmitFunc: t.EmitFunc,
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
	Name string
	Here *tinypkg.Package

	ImportPackages func() ([]*tinypkg.ImportedPackage, error)
	EmitCode       func(w io.Writer) error

	EmitFunc
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
		io.WriteString(w, "\t")
		impkg.Emit(w)
	}
	io.WriteString(w, ")\n")
	return nil
}

var _ Emitter = &Code{}