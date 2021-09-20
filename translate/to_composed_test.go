package translate

import (
	"os"
	"testing"

	"github.com/podhmo/apikit/pkg/tinypkg"
	"github.com/podhmo/apikit/resolve"
)

type Message struct{}
type Greeter struct{}
type Event struct{}

func NewMessage(phrase string) *Message {
	return nil
}
func NewGreeter(m Message) *Greeter {
	return nil
}
func NewEvent(g Greeter) (*Event, error) {
	return nil, nil
}

func TestToComposed(t *testing.T) {
	resolver := resolve.NewResolver()
	here := tinypkg.NewPackage("main", "")

	inputs := []*resolve.Def{
		resolver.Def(NewMessage),
		resolver.Def(NewGreeter),
		resolver.Def(NewEvent),
	}

	w := os.Stdout
	if err := writeComposed(w, here, resolver, "InitializeEvent", inputs); err != nil {
		t.Errorf("unexpected error %+v", err)
	}
}
