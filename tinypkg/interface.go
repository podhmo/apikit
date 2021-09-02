package tinypkg

import "fmt"

type Symboler interface {
	fmt.Stringer
	Symbol() *Symbol
}
type walker interface {
	onWalk(use func(*Symbol) error) error
}
