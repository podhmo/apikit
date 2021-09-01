package resolve

import (
	"reflect"
	"testing"

	"github.com/podhmo/apikit/tinypkg"
)

func TestResolveSymbol(t *testing.T) {
	extractShape := NewResolver().extractor.Extract

	pkg := tinypkg.NewPackage(reflect.TypeOf(DB{}).PkgPath(), "")
	main := tinypkg.NewPackage("main", "main")

	cases := []struct {
		msg    string
		here   *tinypkg.Package
		input  interface{}
		output string
	}{
		{
			// OK: simple symbol
			msg:    "same package symbol",
			here:   pkg,
			input:  DB{},
			output: "DB",
		},
		{
			// OK: imported symbol
			msg:    "imported symbol",
			here:   main,
			input:  DB{},
			output: "resolve.DB",
		},
		{
			// OK: map[<T>, value]
			msg:    "imported map key",
			here:   main,
			input:  map[string]DB{},
			output: "map[string]resolve.DB",
		},
		{
			// OK: map[*<T>, value]
			msg:    "imported map key",
			here:   main,
			input:  map[*DB]int{},
			output: "map[*resolve.DB]int",
		},
		{
			// OK: map[*<T>, value]
			msg:    "imported map key, same package",
			here:   pkg,
			input:  map[*DB]int{},
			output: "map[*DB]int",
		},
		{
			// OK: slice[<T>]
			msg:    "imported slice",
			here:   main,
			input:  []*DB{},
			output: "[]*resolve.DB",
		},
		{
			// OK: nested[<T>]
			msg:    "imported slice",
			here:   main,
			input:  []map[string][]map[*DB][]*DB{},
			output: "[]map[string][]map[*resolve.DB][]*resolve.DB",
		},
		{
			// OK: new type
			msg:    "improted new type",
			here:   main,
			input:  KindComponent,
			output: "resolve.Kind",
		},
		// TODO: func
	}
	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			s := extractShape(c.input)
			importedSymbol := ExtractSymbol(c.here, s)
			if want, got := c.output, importedSymbol.String(); want != got {
				t.Errorf("want:\n\t%q\nbut got:\n\t%q\n", want, got)
			}
		})
	}
}