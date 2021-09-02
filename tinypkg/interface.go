package tinypkg

import "fmt"

type Symboler interface {
	fmt.Stringer
	Symbol() *Symbol
}
type walkerNode interface {
	onWalk(use func(*Symbol) error) error
}
