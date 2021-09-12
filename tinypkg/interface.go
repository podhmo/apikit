package tinypkg

import "fmt"

type Symboler interface {
	fmt.Stringer
	walkerNode
}
type walkerNode interface {
	onWalk(use func(*Symbol) error) error
}
