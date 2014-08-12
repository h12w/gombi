package parse

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hailiang/gspec/core"
	exp "github.com/hailiang/gspec/expectation"
	"github.com/hailiang/gspec/suite"
)

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

var _ = suite.Add(func(s core.S) {
	describe, given, it := suite.Alias3("describe", "given", "it", s)
	expect := exp.Alias(s.FailNow)

	describe("the parser", func() {
		given("a simple arithmetic grammar & sample input tokens", func() {
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

			scanner := newTestScanner([]token{
				{"2", T},
				{"+", Plus},
				{"3", T},
				{"*", Mult},
				{"4", T},
				{"", ruleEOF},
			})

			it("can parses the tokens and generate a correct parse tree", func() {
				ctx := newContext(P)

				for scanner.scan() {
					ctx.cur.each(func(s *state) {
						ctx.scanPredict(s, newTermState(scanner.token()))
					})
					//fmt.Printf("set -> %s\n", ctx.cur.String())
					ctx.shift()
				}

				output := "\n"
				for _, s := range ctx.cur.a {
					s.traverse(0, func(s *state, level int) {
						output += fmt.Sprintf("%s%s\n", strings.Repeat("    ", level), s.expr())
					})
				}
				expect(output).Equal(`
P ::= S EOF•
    S ::= S + M•
        S ::= M•
            M ::= T•
                T ::= 2
        + ::= +
        M ::= M * T•
            M ::= T•
                T ::= 3
            * ::= *
            T ::= 4
    EOF ::= 
`)
			})
		})
	})
})

func TestAll(t *testing.T) {
	suite.Test(t)
}
