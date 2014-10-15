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
	if n == nil {
		return nil
	}
	return n.state.Alt.R
}

func (n *Node) Child(i int) *Node {
	if n == nil {
		return nil
	}
	return newNode(n.values[i])
}

func (n *Node) LastChild() *Node {
	return n.Child(n.ChildCount() - 1)
}

func (n *Node) ChildCount() int {
	return len(n.values)
}

func (n *Node) ID() int {
	return n.token.ID
}

func (n *Node) Value() []byte {
	return n.token.Value
}

func (n *Node) Pos() int {
	if n == nil {
		return 0
	}
	return n.token.Pos
}

func (n *Node) Is(r *R) bool {
	return n.Rule() == r
}

func (n *Node) EachItem(visit func(*Node)) {
	if n == nil {
		return
	}
	cur := n
Loop:
	for {
		if cur.Alt() == cur.rule().Alts[0] {
			visit(cur)
			break Loop
		} else {
			visit(cur.Child(0))
			cur = cur.Child(1)
		}
	}
}

func (n *Node) Each(visit func(*Node)) {
	cur := n
	for cur != nil {
		visit(cur.Child(0))
		cur = cur.Child(1)
	}
}

func (n *Node) Get(r *R) string {
	s := n.find(r)
	if s == nil {
		return ""
	}
	s = s.leaf()
	if s == nil {
		return ""
	}
	return string(s.token.Value)
}

func (s *state) Find(rule *R) *Node {
	return newNode(s.find(rule))
}

func (s *state) find(rule *R) *state {
	if s.rule() == rule {
		return s
	}
	for _, child := range s.values {
		if child != nil {
			if f := child.find(rule); f != nil {
				return f
			}
		}
	}
	return nil
}
func (s *state) leaf() *state {
	if s.isTerm() {
		return s
	} else if len(s.values) == 1 {
		return s.values[0].leaf()
	}
	return nil
}
