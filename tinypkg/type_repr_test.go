package tinypkg

import (
	"strings"
	"testing"
)

func TestTypeRepr(t *testing.T) {
	main := NewPackage("main", "")
	pkg := NewPackage("m/pkg/foo", "")

	cases := []struct {
		msg   string
		here  *Package
		input Symboler
		want  string
	}{
		{
			msg:   "symbol, another package",
			here:  main,
			input: pkg.NewSymbol("Foo"),
			want:  "foo.Foo",
		},
		{
			msg:   "pointer, another package",
			here:  main,
			input: &Pointer{Lv: 1, V: pkg.NewSymbol("Foo")},
			want:  "*foo.Foo",
		},
		{
			msg:   "map, another package",
			here:  main,
			input: &Map{K: builtins.NewSymbol("string"), V: &Pointer{Lv: 1, V: pkg.NewSymbol("Foo")}},
			want:  "map[string]*foo.Foo",
		},
		// TODO: more tests
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			got := ToRelativeTypeString(c.here, c.input)
			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(got); want != got {
				t.Errorf("want:\n\t%q\nbut got:\n\t%q", want, got)
			}
		})
	}
}
