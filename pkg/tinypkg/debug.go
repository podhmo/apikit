package tinypkg

import (
	"fmt"
	"os"
)

func Debug(node Node) {
	// TODO: caller position
	fmt.Fprintf(os.Stdout, "node Type=%T, expr=%s\n", node, node)
}
