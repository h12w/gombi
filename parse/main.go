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

	s0 := newStateSet()
	s0.add(&state{P, P.alts[0], 0, 0})
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
		ss = append(ss, newStateSet())
		ss[i].each(func(s state) {
			if s.complete() { // complete
				ss[s.i].each(func(os state) {
					if os.expect(s.rule) {
						os.d++
						ss[i].add(&os)
					}
				})
			} else if s.expect(scanner.token().rule) { // scan
				s.d++
				ss[i+1].add(&s)
			} else { // predict
				next := s.next()
				for _, alt := range next.alts {
					ss[i].add(&state{next, alt, i, 0})
				}
			}
		})
		fmt.Println("r:", ss[i].expr())
		i++
	}
}
