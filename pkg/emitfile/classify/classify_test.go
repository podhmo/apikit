package classify

import (
	"fmt"
	"strings"
	"testing"
)

func TestCommit(t *testing.T) {
	lines := func(xs ...string) []string { return xs }
	cases := []struct {
		prev []string // "<name>@<content>"
		cur  []string // "<name>@<content>"
		want []string // "<Result>:<name>"
	}{
		{prev: nil, cur: nil, want: nil},
		{prev: nil, cur: lines("hello@"), want: lines("C:hello")},                            // create
		{prev: lines("hello@1"), cur: lines("hello@1"), want: lines("-:hello")},              // not modified
		{prev: lines("hello@1"), cur: lines("hello@2"), want: lines("U:hello")},              // update
		{prev: lines("hello@"), cur: nil, want: lines("D:hello")},                            // deleted
		{prev: lines("hello@1"), cur: lines("hello2@2"), want: lines("C:hello2", "D:hello")}, // deleted (renamed)
	}

	for i, c := range cases {
		c := c

		load := func(k string, xs []string) []Entry {
			entries := make([]Entry, len(xs))
			for i, x := range xs {
				parts := strings.SplitN(x, "@", 2)
				name := parts[0]
				hash := []byte(parts[1])
				entries[i] = NewEntry(name, func() ([]byte, error) {
					return hash, nil
				})
			}
			return entries
		}

		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			prev := load("P", c.prev)
			current := load("C", c.cur)
			results, err := Classify(prev, current)
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
				return
			}

			got := make([]string, len(results))
			for i, r := range results {
				got[i] = fmt.Sprintf("%s:%s", r.Type, r.Name())
			}
			if want, got := strings.TrimSpace(strings.Join(c.want, "\n\t")), strings.TrimSpace(strings.Join(got, "\n\t")); want != got {
				t.Errorf("want:\n\t%s\nbut got:\n\t%s", want, got)
				t.Logf("input         : %+v", c.cur)
				t.Logf("previous input: %+v", c.prev)
			}
		})
	}
}
