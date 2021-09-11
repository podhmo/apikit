package translate

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/tinypkg"
)

type Session struct{}
type Todo struct {
	Title string
	Done  bool
}

func AddTodo(session *Session, title string, done bool) (*Todo, error) {
	return nil, nil
}

func TestWriteRunner(t *testing.T) {
	here := tinypkg.NewPackage("main", "")
	resolver := resolve.NewResolver()
	def := resolver.Resolve(AddTodo)

	provider := tinypkg.NewPackage("m/component", "").NewSymbol("Provider")

	var buf bytes.Buffer
	writeRunner(&buf, here, def, "RunAddTodo", &tinypkg.Var{Name: "provider", Symboler: provider})

	got := buf.String()
	fmt.Fprintln(os.Stdout, got)
}
