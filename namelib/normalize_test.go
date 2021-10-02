package namelib_test

import (
	"fmt"
	"testing"

	"github.com/podhmo/apikit/namelib"
)

func TestToNormalizedOK(t *testing.T) {
	cases := []struct {
		left  string
		right string
	}{
		{
			left: "a", right: "a",
		},
		{
			left: "A", right: "a",
		},
		{
			left: "a", right: "A",
		},
		{
			left: "DB", right: "db",
		},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			want := namelib.ToNormalized(c.left)
			got := namelib.ToNormalized(c.right)
			if want != got {
				t.Errorf("want %q == %q\n\t%q -> %q\n\t%q -> %q", c.left, c.right, c.left, want, c.right, got)
			}
		})
	}
}

func TestToNormalizedNG(t *testing.T) {
	cases := []struct {
		left  string
		right string
	}{
		{
			left: "a", right: "b",
		},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			want := namelib.ToNormalized(c.left)
			got := namelib.ToNormalized(c.right)
			if want == got {
				t.Errorf("want %q != %q, but treated as same value", c.left, c.right)
			}
		})
	}
}
