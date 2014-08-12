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

func (rs Rs) expr() string {
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

func (s *state) expr() string {
	if s.value != nil {
		return fmt.Sprintf("%s ::= %s", s.Name, *s.value)
	}
	return fmt.Sprintf("%s ::= %v•%v", s.Name, s.Alt.Rs[:s.d].expr(), s.Alt.Rs[s.d:].expr())
}

func (s *state) traverse(level int, visit func(*state, int)) {
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
