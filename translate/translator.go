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

type Code struct {
	Name    string
	Imports func(here *tinypkg.Package) []*tinypkg.ImportedPackage
	Emit    func(here *tinypkg.Package, w io.Writer) error
}

func (t *Translator) TranslateInterface(here *tinypkg.Package, name string) *Code {
	return &Code{
		Name: name,
		Imports: func(here *tinypkg.Package) []*tinypkg.ImportedPackage {
			imports := make([]*tinypkg.ImportedPackage, 0, len(t.Tracker.Needs))
			for _, need := range t.Tracker.Needs {

			}
			return imports
		},
		Emit: func(here *tinypkg.Package, w io.Writer) error {
			writeInterface(w, here, t.Tracker, name)
			return nil
		},
	}
}
