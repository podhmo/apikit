package web

import (
	"fmt"
	"log"

	"github.com/podhmo/apikit/web/pathpattern"
)

type T = interface{}
type Node = pathpattern.Node
type Router struct {
	Root         *Node
	ErrorHandler func(error)

	Parent *Router // []*Router?

	FullPrefix string
	prefix     string
}

func NewRouter() *Router {
	return &Router{
		Root: &Node{},
		ErrorHandler: func(err error) {
			log.Printf("ERROR: %+v", err)
		},
	}
}

func (r *Router) Method(method, pattern string, fn T) *Node {
	path := fmt.Sprintf("%s %s%s", method, r.FullPrefix, pattern)
	node, err := r.Root.CreateNode(path, nil)
	if err != nil {
		r.ErrorHandler(err)
		return nil
	}
	node.Value = fn
	return node
}

func (r *Router) Group(pattern string, use func(*Router)) *Router {
	fullprefix := r.FullPrefix + pattern
	child := &Router{Root: r.Root, ErrorHandler: r.ErrorHandler, Parent: r, prefix: pattern, FullPrefix: fullprefix}
	use(child)
	return child
}
