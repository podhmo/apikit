package resolve

import (
	"context"
	"strings"
	"testing"

	"github.com/podhmo/apikit/pkg/difftest"
	"github.com/podhmo/apikit/pkg/tinypkg"
)

type Store struct{}
type Foo struct{}

func UseFoo(store *Store, foo *Foo) error { return nil }

type Bar struct{}

func UseBar(ctx context.Context, store *Store, bar *Bar) error { return nil }

func TestTrackerExtractInterface(t *testing.T) {
	resolver := NewResolver()
	tracker := NewTracker(resolver)

	tracker.Track(resolver.Def(UseFoo))
	tracker.Track(resolver.Def(UseBar))

	// define interface in m/component package
	pkg := resolver.NewPackage("m/component", "")
	resolvePkg := resolver.NewPackageFromInterface(&DB{}, "")

	cases := []struct {
		msg  string
		here *tinypkg.Package
		want string
	}{
		{
			msg:  "another-package",
			here: pkg,
			want: `
type Provider interface {
	Store() *resolve.Store
	Foo() *resolve.Foo
	Bar() *resolve.Bar
}
			`,
		},
		{
			msg:  "same-package",
			here: resolvePkg,
			want: `
type Provider interface {
	Store() *Store
	Foo() *Foo
	Bar() *Bar
}
			`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			iface := tracker.ExtractInterface(c.here, resolver, "Provider")
			var buf strings.Builder
			tinypkg.WriteInterface(&buf, c.here, "Provider", iface) // or WriteInterface(&buf, nil, "", iface)

			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(buf.String()); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}
