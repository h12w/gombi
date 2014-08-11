package main

import "fmt"

type token struct {
	value string
	*rule
}

type testScanner struct {
	tokens []token
	i      int
}

func newTestScanner(tokens []token) *testScanner {
	return &testScanner{tokens: tokens, i: -1}
}

func (s *testScanner) scan() bool {
	s.i++
	return s.i < len(s.tokens)
}

func (s *testScanner) token() *token {
	return &s.tokens[s.i]
}

func main() {
	T := term("T")
	Plus := term("+")
	Mult := term("*")

	M := &rule{name: "M"}
	M.alts = alts{
		{rules{M, Mult, T}},
		{rules{T}},
	}

	S := &rule{name: "S"}
	S.alts = alts{
		{rules{S, Plus, M}},
		{rules{M}},
	}

	EOF := term("EOF")

	P := &rule{name: "P", alts: alts{{rules{S, EOF}}}}

	s0 := &stateSet{}
	s0.add(newState(P, 0))
	ss := []*stateSet{s0}

	scanner := newTestScanner([]token{
		{"2", T},
		{"+", Plus},
		{"3", T},
		{"*", Mult},
		{"4", T},
		{"", EOF},
	})

	i := 0
	for scanner.scan() {
		ss = append(ss, &stateSet{})
		ss[i].each(func(s *state) {
			scanPredict(scanner.token(), ss, s, i)
		})
		fmt.Printf("set(%d) -> %s\n", i, ss[i].expr())
		i++
	}

	ss[i].a[0].traverse(func(s *state) {
		fmt.Println(s.expr())
	})
}

func scanPredict(token *token, ss []*stateSet, s *state, i int) {
	if !s.complete() {
		if ns, ok := s.scan(token); ok { // scan
			if !ns.complete() || ns.rule.alts[0].last().name == "EOF" {
				ss[i+1].add(ns)
			}
			complete(ss, ns, i+1) // complete
		} else { // predict
			predict(token, ss, s, i)
		}
	}
}

func complete(ss []*stateSet, s *state, i int) {
	if s.complete() {
		s.parents.each(func(parent *state) {
			ns := parent.copy()
			ns.addChild(s)
			ns.d++
			if !ns.complete() || ns.rule.alts[0].last().name == "EOF" {
				ss[i].add(ns)
			}
			complete(ss, ns, i)
		})
	}
}

func predict(token *token, ss []*stateSet, s *state, i int) {
	next := s.next()
	for j := range next.alts {
		ss[i].add(newState(next, j)).addParent(s)
	}
}
