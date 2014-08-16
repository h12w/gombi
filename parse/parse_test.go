package parse

import (
	"strings"
	"testing"

	"github.com/hailiang/gspec/core"
	exp "github.com/hailiang/gspec/expectation"
	"github.com/hailiang/gspec/suite"
)

type testScanner struct {
	tokens []*testToken
	i      int
}

func newTestScanner(tokens []*testToken) *testScanner {
	return &testScanner{tokens: tokens, i: -1}
}

func (s *testScanner) Scan() bool {
	s.i++
	return s.i < len(s.tokens)
}

func (s *testScanner) Token() (*Token, *R) {
	t := s.tokens[s.i]
	return t.t, t.r
}

var _ = suite.Add(func(s core.S) {
	describe, testcase, given := suite.Alias3("describe", "testcase:", "given", s)

	describe("the parser", func() {

		given("simple arithmetic grammar", func() {
			var (
				T    = Term("T")
				Plus = Term("+")
				Mult = Term("*")
				M    = Rule("M", Or(
					T,
					Con(Self, Mult, T),
				))
				S = Rule("S", Or(
					Con(Self, Plus, M),
					M,
				))
				P = Rule("P", S, EOF)
			)
			testcase("assotitivity", func() {
				testParse(s, P, TT{
					tok("1", T),
					tok("+", Plus),
					tok("2", T),
					tok("+", Plus),
					tok("3", T),
				}, `
			P ::= S EOF•
				S ::= S + M•
					S ::= S + M•
						S ::= M•
							M ::= T•
								T ::= 1•
						+ ::= +•
						M ::= T•
							T ::= 2•
					+ ::= +•
					M ::= T•
						T ::= 3•
				EOF ::= •`)
			})
			testcase("precedence", func() {
				testParse(s, P, TT{
					tok("2", T),
					tok("+", Plus),
					tok("3", T),
					tok("*", Mult),
					tok("4", T),
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
		})

		given("a grammar with nullable rule", func() {
			var (
				A = Term("A")
				B = Term("B")
				X = Rule("X", B.ZeroOrOne())
				C = Term("C")
				P = Rule("P", Con(A, X).As("AX"), C)
			)

			testcase("a sequence without the optional token", func() {
				testParse(s, P, TT{
					tok("A", A),
					tok("C", C),
				}, `
				P ::= AX C EOF•
					AX ::= A X•
						A ::= A•
					C ::= C•
					EOF ::= •`,
				)
			})

			testcase("a sequence with the optional token", func() {
				testParse(s, P, TT{
					tok("A", A),
					tok("B", B),
					tok("C", C),
				}, `
				P ::= AX C EOF•
					AX ::= A X•
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
			testParse(s, P, TT{
				tok("A", A),
				tok("C", C),
			}, `
			P ::= A Null C EOF•
				A ::= A•
				C ::= C•
				EOF ::= •`)
		})

		given("a grammar with zero or more repetition", func() {
			var (
				A = Term("A")
				B = Term("B")
				X = Rule("X", B.ZeroOrMore())
				C = Term("C")
				P = Rule("P", A, X, C)
			)
			testcase("zero", func() {
				testParse(s, P, TT{
					tok("A", A),
					tok("C", C),
				}, `
			P ::= A X C EOF•
				A ::= A•
				C ::= C•
				EOF ::= •`)
			})

			testcase("one", func() {
				testParse(s, P, TT{
					tok("A", A),
					tok("B", B),
					tok("C", C),
				}, `
			P ::= A X C EOF•
				A ::= A•
				X ::= X B•
					B ::= B•
				C ::= C•
				EOF ::= •`)
			})

			testcase("two", func() {
				testParse(s, P, TT{
					tok("A", A),
					tok("B", B),
					tok("B", B),
					tok("C", C),
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

		given("a grammar with common prefix", func() {
			var (
				A = Term("A")
				B = Term("B")
				X = Rule("X", A)
				Y = Rule("Y", A, B)

				P = Rule("P", Or(X, Y).As("S"))
			)
			testcase("short", func() {
				testParse(s, P, TT{
					tok("A", A),
				}, `
			P ::= X EOF•
				X ::= A•
				EOF ::= •`)
			})
			testcase("short", func() {
				testParse(s, P, TT{
					tok("A", A),
					tok("B", B),
				}, `
			P ::= X B EOF•
				X ::= A•
				B ::= B•
				EOF ::= •`)
			})

		})
	})
})

func TestAll(t *testing.T) {
	suite.Test(t)
}

func testParse(s core.S, P *R, tokens TT, expected string) {
	expect := exp.Alias(s.FailNow, 1)
	scanner := newTestScanner(append(tokens, tok("", EOF)))
	parser := NewParser(P)
	for scanner.Scan() {
		parser.Parse(scanner.Token())
	}
	results := parser.Results()
	expect(len(results)).Equal(1)
	expect(results[0].String()).Equal(unindent(expected + "\n"))
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

type testToken struct {
	t *Token
	r *R
}
type TT []*testToken

func tok(v string, r *R) *testToken {
	return &testToken{&Token{[]byte(v), 0}, r}
}
