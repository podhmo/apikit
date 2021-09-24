package translate

import (
	"reflect"
	"strings"
	"testing"

	"github.com/podhmo/apikit/pkg/difftest"
	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

type Message struct{}
type Greeter struct{}
type Event struct{}

func NewMessage(phrase string) *Message {
	return nil
}
func NewGreeter(m Message) (*Greeter, error) {
	return nil, nil
}
func NewEvent(g Greeter) (*Event, error) {
	return nil, nil
}

func TestToComposed(t *testing.T) {
	resolver := resolve.NewResolver()
	main := resolver.NewPackage("main", "")
	pkg := resolver.NewPackage(reflect.TypeOf(DB{}).PkgPath(), "")

	cases := []struct {
		msg   string
		here  *tinypkg.Package
		input []interface{}
		want  string
	}{
		{
			msg:  "another-package",
			here: main,
			input: []interface{}{
				NewMessage,
				NewGreeter,
				NewEvent,
			},
			want: `
func InitializeEvent(phrase string) (*translate.Event, error) {
	var v0 *translate.Message
	{
		v0 = NewMessage(phrase)
	}
	var v1 *translate.Greeter
	{
		var err error
		v1, err = NewGreeter(m)
		if err != nil {
			return nil, err
		}
	}
	return translate.NewEvent(v1)
}`,
		},
		{
			msg:  "same-package",
			here: pkg,
			input: []interface{}{
				NewMessage,
				NewGreeter,
				NewEvent,
			},
			want: `
func InitializeEvent(phrase string) (*Event, error) {
	var v0 *Message
	{
		v0 = NewMessage(phrase)
	}
	var v1 *Greeter
	{
		var err error
		v1, err = NewGreeter(m)
		if err != nil {
			return nil, err
		}
	}
	return NewEvent(v1)
}`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			providers := make([]*resolve.Def, len(c.input))
			for i, fn := range c.input {
				providers[i] = resolver.Def(fn)
			}

			var buf strings.Builder
			if err := writeComposed(&buf, c.here, resolver, "InitializeEvent", providers); err != nil {
				t.Errorf("unexpected error %+v", err)
			}
			if want, got := strings.TrimSpace(c.want), strings.TrimSpace(buf.String()); want != got {
				difftest.LogDiffGotStringAndWantString(t, got, want)
			}
		})
	}
}
