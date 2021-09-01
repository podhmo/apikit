package translate

import (
	"reflect"
	"strings"
	"testing"

	"github.com/podhmo/apikit/difftest"
	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/tinypkg"
)

type DB struct{}
type User struct{}

func ListUser(db *DB) []User { return nil }

func TestTracker(t *testing.T) {
	main := tinypkg.NewPackage("main", "main")
	pkg := tinypkg.NewPackage(reflect.TypeOf(DB{}).PkgPath(), "")
	resolver := resolve.NewResolver()

	cases := []struct {
		msg   string
		input []interface{}
		here  *tinypkg.Package
		want  string
	}{
		{
			msg:   "1 component, another package",
			here:  main,
			input: []interface{}{ListUser},
			want: `
type Component interface {
	Db() *translate.DB
}`,
		},
		{
			msg:   "1 component, same package",
			here:  pkg,
			input: []interface{}{ListUser},
			want: `
type Component interface {
	Db() *DB
}`,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			tracker := NewTracker()
			for _, x := range c.input {
				def := resolver.Resolve(x)
				tracker.Track(def)
			}

			buf := new(strings.Builder)
			WriteInterface(buf, c.here, tracker, "Component")
			got := buf.String()
			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(got); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}
