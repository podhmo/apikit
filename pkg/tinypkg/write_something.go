package tinypkg

import (
	"fmt"
	"io"
	"strings"
)

func WriteFunc(w io.Writer, here *Package, name string, f *Func, body func() error) error {
	if name == "" {
		name = f.Name
	}

	args := make([]string, 0, len(f.Args))
	for _, x := range f.Args {
		args = append(args, ToRelativeTypeString(here, x))
	}

	returns := make([]string, 0, len(f.Returns))
	for _, x := range f.Returns {
		returns = append(returns, ToRelativeTypeString(here, x))
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

func WriteInterface(w io.Writer, here *Package, name string, iface *Interface) error {
	if name == "" {
		name = iface.Name
	}
	// interface <name> {
	// ..
	//}
	fmt.Fprintf(w, "type %s interface {\n", name)
	defer fmt.Fprintln(w, "}")
	for _, method := range iface.Methods {
		fmt.Fprintf(w, "\t%s\n", ToInterfaceMethodString(here, method.Name, method))
	}
	return nil
}
