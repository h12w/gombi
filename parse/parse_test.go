package parse

import (
	"fmt"
	"testing"
)

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

func TestParse(t *testing.T) {
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

	P := &rule{name: "P", alts: alts{{rules{S, ruleEOF}}}}

	ctx := newContext(P)

	scanner := newTestScanner([]token{
		{"2", T},
		{"+", Plus},
		{"3", T},
		{"*", Mult},
		{"4", T},
		{"", ruleEOF},
	})

	for scanner.scan() {
		ctx.cur.each(func(s *state) {
			ctx.scanPredict(scanner.token(), s)
		})
		fmt.Printf("set -> %s\n", ctx.cur.expr())
		ctx.shift()
	}

	ctx.cur.a[0].traverse(func(s *state) {
		fmt.Println(s.expr())
	})
}
