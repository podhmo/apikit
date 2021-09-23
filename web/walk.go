package web

import (
	"fmt"

	"github.com/podhmo/apikit/web/pathpattern"
)

func Walk(r *Router, use func(*WalkerNode) error) error {
	return walk(r.Root, use, nil)
}

func walk(node *Node, use func(*WalkerNode) error, history []pathpattern.Suffix) error {
	if node.Value != nil {
		if err := use(&WalkerNode{Node: node, History: history}); err != nil {
			return err
		}
	}

	for _, suffix := range node.Suffixes {
		if err := walk(suffix.Node, use, append(history, suffix)); err != nil {
			return err
		}
	}
	return nil
}

type WalkerNode struct {
	Node    *Node
	History []pathpattern.Suffix
}

func (n *WalkerNode) Path() []string {
	parts := make([]string, 0, len(n.History))
	i := 0
	for _, suffix := range n.History {
		if suffix.Pattern != "" {
			parts = append(parts, suffix.Pattern)
		}

		if len(suffix.Node.VariableNames) > i {
			parts = append(parts, fmt.Sprintf("{%s}", suffix.Node.VariableNames[i]))
			i++
		}
	}
	return parts
}
