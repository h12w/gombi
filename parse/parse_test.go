package parse

import (
	"testing"

	"github.com/hailiang/gspec/core"
	exp "github.com/hailiang/gspec/expectation"
	"github.com/hailiang/gspec/suite"
)

type testScanner struct {
	tokens []*Token
	i      int
}

func newTestScanner(tokens []*Token) *testScanner {
	return &testScanner{tokens: tokens, i: -1}
}

func (s *testScanner) Scan() bool {
	s.i++
	return s.i < len(s.tokens)
}

func (s *testScanner) Token() *Token {
	return s.tokens[s.i]
}

var _ = suite.Add(func(s core.S) {
	describe, given, it := suite.Alias3("describe", "given", "it", s)
	expect := exp.Alias(s.FailNow)

	describe("the parser", func() {

		given("a parser of simple arithmetic grammar & input tokens", func() {
			T := Rule("T")
			Plus := Rule("+")
			Mult := Rule("*")
			M := Rule("M")
			M.Or(
				Con(M, Mult, T),
				T,
			)
			S := Rule("S")
			S.Or(
				Con(S, Plus, M),
				M,
			)
			P := Rule("P").Con(S, EOF)
			parser := NewParser(P)

			scanner := newTestScanner([]*Token{
				{"2", T},
				{"+", Plus},
				{"3", T},
				{"*", Mult},
				{"4", T},
				{"", EOF},
			})

			it("can parse the tokens and generate a correct parse tree", func() {
				for scanner.Scan() {
					parser.Parse(scanner.Token())
				}
				result := parser.Result()
				expect(result).NotEqual(nil)
				expect(result.String()).Equal(`
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

		given("a grammar with nullable rule", func() {
			A := Rule("A")
			B := Rule("B")
			X := Rule("X").Or(B, Null)
			C := Rule("C")

			P := Rule("P").Con(A, X, C, EOF)

			given("a sequence without the optional token", func() {
				scanner := newTestScanner([]*Token{
					{"A", A},
					{"C", C},
					{"EOF", EOF},
				})
				parser := NewParser(P)
				for scanner.Scan() {
					parser.Parse(scanner.Token())
				}
				result := parser.Result()
				expect(result).NotEqual(nil)
				expect(result.String()).Equal(`
P ::= A X C EOF•
    A ::= A
    C ::= C
    EOF ::= EOF
`)
			})

			given("a sequence with the optional token", func() {
				scanner := newTestScanner([]*Token{
					{"A", A},
					{"B", B},
					{"C", C},
					{"EOF", EOF},
				})
				parser := NewParser(P)
				for scanner.Scan() {
					parser.Parse(scanner.Token())
				}
				result := parser.Result()
				expect(result).NotEqual(nil)
				expect(result.String()).Equal(`
P ::= A X C EOF•
    A ::= A
    X ::= B•
        B ::= B
    C ::= C
    EOF ::= EOF
`)
			})

		})
	})
})

func TestAll(t *testing.T) {
	suite.Test(t)
}
