package main

import (
	"fmt"
	"strings"
)

type state struct {
	*rule
	*alt
	d        int // dot position
	parents  stateSet
	children []*state
	value    *string
}

func newState(r *rule, altIndex int) *state {
	alt := r.alts[altIndex]
	return &state{rule: r, alt: alt, children: make([]*state, len(alt.rules))}
}

func newTermState(t *token) *state {
	return &state{rule: t.rule, value: &t.value}
}

func (s *state) expr() string {
	if s.value != nil {
		return fmt.Sprintf("%s ::= %s", s.name, *s.value)
	}
	return fmt.Sprintf("%s ::= %v•%v", s.name, s.alt.rules[:s.d].expr(), s.alt.rules[s.d:].expr())
}

func (s *state) addChild(child *state) {
	s.children[s.d] = child
}

func (s *state) addParent(parent *state) {
	s.parents.add(parent)
}

func (s *state) copy() *state {
	c := *s
	c.children = append([]*state{}, s.children...)
	return &c
}

func (s *state) equal(o *state) bool {
	return s.alt == o.alt && s.d == o.d
}

func (s *state) complete() bool {
	return s.d == len(s.alt.rules)
}

func (s *state) next() *rule {
	return s.alt.rules[s.d]
}

func (s *state) scan(t *token) (*state, bool) {
	if s.expect(t.rule) {
		ns := s.copy()
		ns.addChild(newTermState(t))
		ns.d++
		return ns, true
	}
	return nil, false
}

func (s *state) expect(r *rule) bool {
	return !s.complete() && s.alt.rules[s.d] == r
}

func (s *state) traverse(visit func(*state)) {
	visit(s)
	for _, c := range s.children {
		c.traverse(visit)
	}
}

type stateSet struct {
	a []*state
}

func (set *stateSet) empty() bool {
	return len(set.a) == 0
}

func (set *stateSet) add(o *state) *state {
	for _, s := range set.a {
		if s.equal(o) {
			return s
		}
	}
	set.a = append(set.a, o)
	return o
}

func (set *stateSet) each(visit func(*state)) {
	i := 0
	for {
		if i == len(set.a) {
			break
		}
		visit(set.a[i])
		i++
	}
}

func (set *stateSet) next() func() *state {
	i := -1
	return func() *state {
		i++
		if i < len(set.a) {
			return set.a[i]
		}
		return nil
	}
}

func (s *stateSet) size() int {
	return len(s.a)
}

func (s *stateSet) expr() string {
	ss := make([]string, 0, len(s.a))
	for _, state := range s.a {
		ss = append(ss, state.expr())
	}
	return strings.Join(ss, ", ")
}
