package parse

import (
	"strings"
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
	describe, testcase := suite.Alias2("describe", "testcase:", s)

	describe("the parser", func() {

		testcase("simple arithmetic grammar", func() {
			var (
				T    = Term("T")
				Plus = Term("+")
				Mult = Term("*")
				M    = Rule("M", Or(
					Con(Self, Mult, T),
					T,
				))
				S = Rule("S", Or(
					Con(Self, Plus, M),
					M,
				))
				P = Rule("P", S)
			)
			testParse(s, P, []*Token{
				{"2", T},
				{"+", Plus},
				{"3", T},
				{"*", Mult},
				{"4", T},
			}, `
			P ::= S EOF•
				S ::= S + M•
					S ::= M•
						M ::= T•
							T ::= 2•
					+ ::= +•
					M ::= M * T•
						M ::= T•
							T ::= 3•
						* ::= *•
						T ::= 4•
				EOF ::= •`)
		})

		describe("a parser with nullable rule", func() {
			var (
				A = Term("A")
				B = Term("B")
				X = Rule("X", B.ZeroOrOne())
				C = Term("C")
				P = Rule("P", A, X, C)
			)

			testcase("a sequence without the optional token", func() {
				testParse(s, P, []*Token{
					{"A", A},
					{"C", C},
				}, `
				P ::= A X C EOF•
					A ::= A•
					C ::= C•
					EOF ::= •`,
				)
			})

			testcase("a sequence with the optional token", func() {
				testParse(s, P, []*Token{
					{"A", A},
					{"B", B},
					{"C", C},
				}, `
				P ::= A X C EOF•
					A ::= A•
					X ::= B•
						B ::= B•
					C ::= C•
					EOF ::= •`)
			})
		})

		testcase("a trivial but valid nullable rule", func() {
			var (
				A = Term("A")
				C = Term("C")
				P = Rule("P", A, Null, C)
			)
			testParse(s, P, []*Token{
				{"A", A},
				{"C", C},
			}, `
			P ::= A Null C EOF•
				A ::= A•
				C ::= C•
				EOF ::= •`)
		})

		testcase("a case of zero or more repetition", func() {
			var (
				A = Term("A")
				B = Term("B")
				X = Rule("X", B.ZeroOrMore())
				C = Term("C")
				P = Rule("P", A, X, C)
			)
			testParse(s, P, []*Token{
				{"A", A},
				{"C", C},
			}, `
			P ::= A X C EOF•
				A ::= A•
				C ::= C•
				EOF ::= •`)

			testParse(s, P, []*Token{
				{"A", A},
				{"B", B},
				{"C", C},
			}, `
			P ::= A X C EOF•
				A ::= A•
				X ::= X B•
					B ::= B•
				C ::= C•
				EOF ::= •`)

			testParse(s, P, []*Token{
				{"A", A},
				{"B", B},
				{"B", B},
				{"C", C},
			}, `
			P ::= A X C EOF•
				A ::= A•
				X ::= X B•
					X ::= X B•
						B ::= B•
					B ::= B•
				C ::= C•
				EOF ::= •`)
		})
	})
})

func TestAll(t *testing.T) {
	suite.Test(t)
}

func testParse(s core.S, P *R, tokens []*Token, expected string) {
	expect := exp.Alias(s.FailNow, 1)
	scanner := newTestScanner(append(tokens, &Token{"", EOF}))
	parser := NewParser(P)
	for scanner.Scan() {
		parser.Parse(scanner.Token())
	}
	result := parser.Result()
	expect(result).NotEqual(nil)
	expect(result.String()).Equal(unindent(expected))
}

func unindent(s string) string {
	lines := strings.Split(s, "\n")
	indent := ""
	done := false
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			for _, r := range line {
				if r == ' ' || r == '\t' {
					indent += string(r)
				} else {
					done = true
					break
				}
			}
		}
		if done {
			break
		}
	}
	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], indent)
	}
	return strings.Join(lines, "\n")
}
