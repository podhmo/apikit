package tinypkg

import (
	"fmt"
	"io"
	"strings"
)

func WriteFunc(w io.Writer, name string, f *Func, body func() error) error {
	args := make([]string, 0, len(f.Args))
	for _, x := range f.Args {
		args = append(args, x.String())
	}

	returns := make([]string, 0, len(f.Returns))
	for _, x := range f.Returns {
		returns = append(returns, x.String())
	}

	// func <name>(<args>...) (<returns>) {
	// ...
	// }
	defer fmt.Fprintln(w, "}")
	switch len(returns) {
	case 0:
		fmt.Fprintf(w, "func %s(%s) {\n", name, strings.Join(args, ", "))
	case 1:
		fmt.Fprintf(w, "func %s(%s) %s {\n", name, strings.Join(args, ", "), returns[0])
	default:
		fmt.Fprintf(w, "func %s(%s) (%s) {\n", name, strings.Join(args, ", "), strings.Join(returns, ", "))
	}
	return body()
}
