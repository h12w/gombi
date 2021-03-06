package scan

import (
	"fmt"
	"strings"
	"testing"

	"h12.io/gspec"

	//	ogdl "github.com/ogdl/flow"
)

const EOF = 0

var _ = gspec.Add(func(s gspec.S) {
	describe, given, it, they, and := gspec.Alias5("describe", "given", "it", "they", "and", s)
	expect := gspec.Expect(s.FailNow)
	describe("patterns", func() {
		given("a pattern", func() {
			a := Pat("ab")
			it("can be repeated zero or one time", func() {
				expect(a.ZeroOrOne().String()).Equal("(?:ab)?")
			})
			it("can be repeated zero or more times", func() {
				expect(a.ZeroOrMore().String()).Equal("(?:ab)*")
			})
			it("can be repeated one or more times", func() {
				expect(a.OneOrMore().String()).Equal("(?:ab)+")
			})
		})
		given("multiple patterns", func() {
			a, b, c := Pat("aa"), Pat("bb"), Pat("cc")
			they("can be concatenated into one pattern", func() {
				one := Con(a, b, c)
				expect(one.String()).Equal("aabbcc")
			})
			they("can be alternative choices", func() {
				alt := Or(a, b, c)
				expect(alt.String()).Equal("aa|bb|cc")
			})
		})
	})
	describe("character sets", func() {
		given("a character set", func() {
			c := Char(`abcde`)
			it("can be converted to a canonical pattern", func() {
				expect(c.String()).Equal(`[a-e]`)
			})
			it("can be negated", func() {
				nc := c.Negate()
				expect(nc.String()).Equal(`[^a-e]`)
				and("the original pattern is not changed by the negation", func() {
					expect(c.String()).Equal(`[a-e]`)
				})
				and("the negated pattern can be negated back", func() {
					expect(nc.Negate().String()).Equal(`[a-e]`)
				})
			})
			it("can exclude a subset", func() {
				expect(c.Exclude(Char(`bd`)).String()).Equal(`[ace]`)
			})
		})
		given("a character set with a single element (OpLiteral)", func() {
			a := Char(`a`)
			it("can be negated", func() {
				na := a.Negate()
				expect(na.String()).Equal(`[^a]`)
				expect(na.Negate().String()).Equal(`[a]`)
			})
		})
		given("a character set with an letter of upper and lower case (OpLiteral)", func() {
			x := Char(`xX`)
			it("has a canonical form of character set", func() {
				expect(x.String()).Equal(`[Xx]`)
			})
			it("can be negated", func() {
				nx := x.Negate()
				expect(nx.String()).Equal(`[^Xx]`)
				expect(nx.Negate().String()).Equal(`[Xx]`)
			})
		})
		given("multiple character sets", func() {
			a := Char(`a`)
			mn := Char(`mn`)
			z := Char(`z`)
			they("can be merged into one character set", func() {
				one := Merge(a, mn, z)
				expect(one.String()).Equal(`[am-nz]`)
			})
		})
	})
	describe("the scanner", func() {
		a := Char(`a`)
		s := Char(` `)
		b := Char(`b`)
		m := NewMatcher(a, s, b)
		scanner, _ := NewStringScanner(m, "b a")
		expect(scanner.Scan()).Equal(true)
		expect(scanner.Error()).Equal(nil)
		expect(scanner.Token()).Equal(
			&Token{
				ID:    3,
				Value: []byte("b"),
				Pos:   0,
			})
		expect(scanner.Scan()).Equal(true)
		expect(scanner.Error()).Equal(nil)
		expect(scanner.Token()).Equal(
			&Token{
				ID:    2,
				Value: []byte(" "),
				Pos:   1,
			})
		expect(scanner.Scan()).Equal(true)
		expect(scanner.Error()).Equal(nil)
		expect(scanner.Token()).Equal(
			&Token{
				ID:    1,
				Value: []byte("a"),
				Pos:   2,
			})
		expect(scanner.Scan()).Equal(true)
		expect(scanner.Error()).Equal(nil)
		expect(scanner.Token()).Equal(
			&Token{
				ID:    EOF,
				Value: nil,
				Pos:   3,
			})
		expect(scanner.Scan()).Equal(false)
		expect(scanner.Error()).Equal(nil)
	})
})

func TestAll(t *testing.T) {
	gspec.Test(t)
}

func pp(v ...interface{}) {
	fmt.Println(v...)
}

/*
func op(v interface{}) {
	buf, _ := ogdl.MarshalIndent(v, "    ", "    ")
	typ := ""
	if v != nil {
		typ = reflect.TypeOf(v).String() + "\n"
	}
	pp("\n" +
		typ +
		string(buf) +
		"\n")
}
*/

func NewStringScanner(m *Matcher, text string) (*Scanner, error) {
	s := Scanner{Matcher: m}
	return &s, s.SetReader(strings.NewReader(text))
}
