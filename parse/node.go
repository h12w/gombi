package parse

type Node struct {
	*state
}

func (n *Node) Rule() *Alt {
	return n.Alt
}

func (n *Node) Child(i int) *Node {
	return &Node{n.values[i]}
}

func (n *Node) ChildCount() int {
	return len(n.values)
}

func (n *Node) Value() interface{} {
	return n.value
}
