package parse

type Node struct {
	*state
}

func newNode(s *state) *Node {
	return &Node{s}
}

func (n *Node) Rule() *Alt {
	return n.Alt
}

func (n *Node) Child(i int) *Node {
	return newNode(n.values[i])
}

func (n *Node) ChildCount() int {
	return len(n.values)
}

func (n *Node) Value() []byte {
	return n.token.Value
}

func (n *Node) Pos() int {
	return n.token.Pos
}
