package main

import (
	"fmt"
	"strings"
)

type state struct {
	*rule
	*alt
	i int // input position
	d int // dot position
}

func (s *state) expr() string {
	return fmt.Sprintf("%s(%d) ::= %v•%v", s.name, s.i, s.alt.rules[:s.d].expr(), s.alt.rules[s.d:].expr())
}

func (s *state) complete() bool {
	return s.d == len(s.alt.rules)
}

func (s *state) next() *rule {
	return s.alt.rules[s.d]
}

func (s *state) expect(r *rule) bool {
	return !s.complete() && s.alt.rules[s.d] == r
}

type stateSet struct {
	a []state
}

func newStateSet() *stateSet {
	return &stateSet{}
}

func (set *stateSet) add(state *state) {
	for _, s := range set.a {
		if s == *state {
			return
		}
	}
	set.a = append(set.a, *state)
}

func (set *stateSet) each(visit func(state)) {
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
			return &set.a[i]
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
