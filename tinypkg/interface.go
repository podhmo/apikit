package tinypkg

type Node interface {
	onWalk(use func(*Symbol) error) error
}
