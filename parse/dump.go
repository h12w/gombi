package parse

import (
	"fmt"
	"strings"
)

func (r *R) expr() string {
	return r.Name
}

func (r *R) String() string {
	return fmt.Sprintf("%s ::= %s", r.Name, r.Alts.expr())
}

func (rs Rules) expr() string {
	ss := make([]string, len(rs))
	for i := range rs {
		ss[i] = rs[i].expr()
	}
	return strings.Join(ss, " ")
}

func (as Alts) expr() string {
	ss := make([]string, len(as))
	for i := range as {
		ss[i] = as[i].expr()
	}
	return strings.Join(ss, " | ")
}

func (s *state) name() string {
	return s.Alt.Parent.Name
}

func (s *state) expr() string {
	if s.value != nil {
		if s.d == 0 {
			return fmt.Sprintf("%s ::= •%v", s.name(), s.value)
		} else if s.d == 1 {
			return fmt.Sprintf("%s ::= %v•", s.name(), s.value)
		}
	}
	return fmt.Sprintf("%s ::= %v•%v", s.name(), s.Alt.Rules[:s.d].expr(), s.Alt.Rules[s.d:].expr())
}

func (s *state) traverse(level int, visit func(*state, int)) {
	if s == nil {
		return
	}
	visit(s, level)
	for _, c := range s.values {
		c.traverse(level+1, visit)
	}
}

func (ss *stateSet) String() string {
	strs := make([]string, 0, len(ss.a))
	for _, s := range ss.a {
		strs = append(strs, s.expr())
	}
	return strings.Join(strs, ", ")
}

func (n *Node) String() string {
	output := "\n"
	indent := "\t"
	n.traverse(0, func(s *state, level int) {
		output += fmt.Sprintf("%s%s\n", strings.Repeat(indent, level), s.expr())
	})
	return output
}
