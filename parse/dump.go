package parse

import (
	"fmt"
	"strings"
)

func (r *rule) expr() string {
	return r.name
}

func (r *rule) String() string {
	return fmt.Sprintf("%s ::= %s", r.name, r.alts.expr())
}

func (rs rules) expr() string {
	ss := make([]string, len(rs))
	for i := range rs {
		ss[i] = rs[i].expr()
	}
	return strings.Join(ss, " ")
}

func (as alts) expr() string {
	ss := make([]string, len(as))
	for i := range as {
		ss[i] = as[i].expr()
	}
	return strings.Join(ss, " | ")
}

func (s *state) expr() string {
	if s.value != nil {
		return fmt.Sprintf("%s ::= %s", s.name, *s.value)
	}
	return fmt.Sprintf("%s ::= %v•%v", s.name, s.alt.rules[:s.d].expr(), s.alt.rules[s.d:].expr())
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
