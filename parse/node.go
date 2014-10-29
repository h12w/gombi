package parse

type Node struct {
	alt    *Alt
	token  *Token
	values []*Node
}
type Token struct {
	ID    int
	Value []byte
	Pos   int
}

func (n *Node) copy() *Node {
	if n == nil {
		return nil
	}
	node := *n
	n = &node
	n.values = append([]*Node(nil), n.values...)
	return n
}

func (n *Node) Rule() *R {
	if n == nil {
		return nil
	}
	return n.alt.R
}

func (n *Node) Child(i int) *Node {
	if n == nil {
		return nil
	}
	return n.values[i]
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
		if cur.alt == cur.alt.R.Alts[0] {
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

func (s *Node) Find(rule *R) *Node {
	return s.find(rule)
}

func (s *Node) find(rule *R) *Node {
	if s.alt.R == rule {
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
func (s *Node) leaf() *Node {
	if s.alt.R.isTerm() {
		return s
	} else if len(s.values) == 1 {
		return s.values[0].leaf()
	}
	return nil
}
