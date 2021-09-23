package tinypkg

import (
	"strings"
	"testing"

	"github.com/podhmo/apikit/pkg/difftest"
)

func TestWriteInterface(t *testing.T) {
	pkg := NewPackage("net/http", "")
	main := NewPackage("main", "")

	iface := pkg.NewInterface(
		"Handler",
		[]*Func{
			pkg.NewFunc("ServeHTTP", []*Var{
				{Node: pkg.NewSymbol("ResponseWriter")},
				{Node: &Pointer{Lv: 1, V: pkg.NewSymbol("Request")}},
			}, nil),
		},
	)

	cases := []struct {
		msg   string
		input *Interface
		here  *Package
		want  string
	}{
		{
			msg:   "same package",
			input: iface,
			here:  pkg,
			want: `
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}`,
		},
		{
			msg:   "another package",
			input: iface,
			here:  main,
			want: `
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			var buf strings.Builder
			if err := WriteInterface(&buf, c.here, "", iface); err != nil {
				t.Fatalf("unexpected error %+v", err)
			}
			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(buf.String()); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}
