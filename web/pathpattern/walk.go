package pathpattern

import "fmt"

// added by me

func Walk(node *Node, use func(*WalkerNode) error) error {
	return walk(node, use, nil)
}

func walk(node *Node, use func(*WalkerNode) error, history []Suffix) error {
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
	History []Suffix
}

func (n *WalkerNode) Path() []string {
	parts := make([]string, 0, len(n.History))
	for _, suffix := range n.History {
		if len(suffix.Node.VariableNames) > 0 {
			parts = append(parts, fmt.Sprintf("{%s}", suffix.Node.VariableNames[len(suffix.Node.VariableNames)-1]))
		} else {
			parts = append(parts, suffix.Pattern)
		}
	}
	return parts
}
