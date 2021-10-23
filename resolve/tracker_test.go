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

type Boo struct{}

func UseBoo(boo *Boo) error                                    { return nil }
func NewBoo(store *Store) *Boo                                 { return nil }
func NewBooWithContext(ctx context.Context, store *Store) *Boo { return nil }

func UseBooWithStore(boo *Boo, store *Store) error               { return nil }
func UseBooWithAnotherStore(boo *Boo, anotherStore *Store) error { return nil }

func TestTrackerExtractInterface(t *testing.T) {
	resolver := NewResolver()

	// define interface in m/component package
	pkg := resolver.NewPackage("m/component", "")
	resolvePkg := resolver.NewPackageFromInterface(&DB{}, "")

	cases := []struct {
		msg    string
		here   *tinypkg.Package
		modify func(*Tracker)
		want   string
	}{
		{
			msg:  "empty",
			here: pkg,
			want: `
type Provider interface {
}
			`,
		},
		{
			msg:  "another-package",
			here: pkg,
			modify: func(tracker *Tracker) {
				tracker.Track(resolver.Def(UseFoo))
				tracker.Track(resolver.Def(UseBar))
			},
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
			modify: func(tracker *Tracker) {
				tracker.Track(resolver.Def(UseFoo))
				tracker.Track(resolver.Def(UseBar))
			},
			want: `
type Provider interface {
	Store() *Store
	Foo() *Foo
	Bar() *Bar
}
			`,
		},
		{
			msg:  "with-override",
			here: resolvePkg,
			modify: func(tracker *Tracker) {
				tracker.Track(resolver.Def(UseBoo)) // func(*Boo)
				tracker.Override("", NewBoo)        // func(*Store) *Boo
			},
			want: `
type Provider interface {
	Boo(store *Store) *Boo
	Store() *Store
}
			`,
		},
		{
			msg:  "with-override-with-context",
			here: resolvePkg,
			modify: func(tracker *Tracker) {
				tracker.Track(resolver.Def(UseBoo))     // func(*Boo)
				tracker.Override("", NewBooWithContext) // func(context.Context, *Store) *Boo
			},
			want: `
type Provider interface {
	Boo(ctx context.Context, store *Store) *Boo
	Store() *Store
}
			`,
		},
		{
			msg:  "with-override-no-effect-veersion",
			here: resolvePkg,
			modify: func(tracker *Tracker) {
				tracker.Track(resolver.Def(UseBooWithStore)) // func(*Boo, *Store)
				tracker.Override("", NewBoo)                 // func(*Store) *Boo
			},
			want: `
type Provider interface {
	Boo(store *Store) *Boo
	Store() *Store
}
			`,
		},
		{
			msg:  "with-override-with-another-name",
			here: resolvePkg,
			modify: func(tracker *Tracker) {
				tracker.Track(resolver.Def(UseBooWithAnotherStore)) // func(*Boo, anotherDB *Store)
				tracker.Override("", NewBoo)                        // func(*Store) *Boo
			},
			want: `
type Provider interface {
	Boo(store *Store) *Boo
	AnotherStore() *Store
	Store() *Store
}
			`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			tracker := NewTracker(resolver)
			if c.modify != nil {
				c.modify(tracker)
			}

			iface := tracker.ExtractInterface(c.here, resolver, "Provider")
			var buf strings.Builder
			tinypkg.WriteInterface(&buf, c.here, "Provider", iface) // or WriteInterface(&buf, nil, "", iface)

			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(buf.String()); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}
