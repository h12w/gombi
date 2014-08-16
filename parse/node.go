package parse

type Node struct {
	*state
}

func newNode(s *state) *Node {
	if s == nil {
		return nil
	}
	return &Node{s}
}

func (n *Node) Alt() *Alt {
	return n.state.Alt
}

func (n *Node) Rule() *R {
	return n.state.Alt.Parent
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
