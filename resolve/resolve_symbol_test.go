package resolve

import (
	"reflect"
	"testing"

	"github.com/podhmo/apikit/tinypkg"
)

func TestResolveSymbol(t *testing.T) {
	extract := NewResolver().extractor.Extract

	pkg := tinypkg.NewPackage(reflect.TypeOf(DB{}).PkgPath(), "")
	main := tinypkg.NewPackage("main", "main")

	cases := []struct {
		msg    string
		here   *tinypkg.Package
		input  interface{}
		output string
	}{
		{
			msg:    "same package symbol",
			here:   pkg,
			input:  DB{},
			output: "DB",
		},
		{
			msg:    "imported symbol",
			here:   main,
			input:  DB{},
			output: "resolve.DB",
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			importedSymbol := ExtractSymbol(c.here, extract(DB{}))
			if want, got := c.output, importedSymbol.String(); want != got {
				t.Errorf("want: %q but got %q", want, got)
			}
		})
	}
}
