package dfa

import (
	"fmt"
	"testing"

	"github.com/hailiang/gspec"
)

var (
	s          = Str
	c          = Char
	b          = Between
	con        = Con
	or         = Or
	and        = And
	zeroOrMore = ZeroOrMore
	zeroOrOne  = ZeroOrOne
	oneOrMore  = OneOrMore
	graphOpt   = &GraphOption{"Ubuntu Mono", true}
)

const (
	hexLabel = iota
	decimalLabel
	identLabel
)

var threeToken = func() *Machine {
	decimalDigit := b('0', '9')
	hexDigit := or(b('0', '9'), b('a', 'f'), b('A', 'F'))
	letter := or(b('a', 'z'), b('A', 'Z'))

	hexLit := con(s(`0`), c(`xX`), oneOrMore(hexDigit)).As(hexLabel)
	decimalLit := oneOrMore(decimalDigit).As(decimalLabel)
	ident := con(letter, zeroOrMore(or(letter, decimalDigit))).As(identLabel)

	return or(hexLit, decimalLit, ident)
}()

var bsas = func() *Machine {
	bsa := con(zeroOrMore(s("b")), s("a"))
	return zeroOrMore(bsa, bsa)
}()

var asbs = func() *Machine {
	asb := con(zeroOrMore(s("a")), s("b"))
	return zeroOrMore(asb, asb)
}()

func TestExpr(t *testing.T) {
	expect := gspec.Expect(t.FailNow)
	for i, testcase := range []struct {
		m *Machine
		s string
	}{
		{
			s(""), `
			s0$
		`},
		{
			s("a"), `
			s0
				'a'     s1
			s1$
		`},
		{
			s("ab"), `
			s0
				'a'     s1
			s1
				'b'     s2
			s2$
		`},
		{
			c("abc"), `
			s0
				'a'-'c' s1
			s1$
		`},
		{
			b('a', 'c'),
			c("abc").dump(),
		},
		{
			con(s("a"), s("b")),
			s("ab").dump(),
		},
		{
			or(s("a"), s("b")),
			c("ab").dump(),
		},
		{
			zeroOrMore(s("a")), `
			s0$
				'a'     s0
		`},
		{
			zeroOrMore(s("ab")), `
			s0$
				'a'     s1
			s1
				'b'     s0
		`},
		{
			oneOrMore(s("ab")), `
			s0
				'a'     s1
			s1
				'b'     s2
			s2$
				'a'     s1
		`},
		{
			threeToken, `
			s0
				'0'     s1
				'1'-'9' s2
				'A'-'Z' s3
				'a'-'z' s3
			s1$3
				'0'-'9' s2
				'X'     s4
				'x'     s4
			s2$3
				'0'-'9' s2
			s3$4
				'0'-'9' s3
				'A'-'Z' s3
				'a'-'z' s3
			s4
				'0'-'9' s5
				'A'-'F' s5
				'a'-'f' s5
			s5$2
				'0'-'9' s5
				'A'-'F' s5
				'a'-'f' s5
		`},
		{
			b(0, '\U0010ffff'), `
			s0
				00-7f   s1
				c2-df   s2
				e0      s3
				e1-ef   s4
				f0      s5
				f1-f3   s6
				f4      s7
			s1$
			s2
				80-bf   s1
			s3
				a0-bf   s2
			s4
				80-bf   s2
			s5
				90-bf   s4
			s6
				80-bf   s4
			s7
				80-8f   s4
			`,
		},
		{
			bsas, `
			s0$
				'a'     s1
				'b'     s0
			s1
				'a'     s0
				'b'     s1
		`},
		{
			bsas.Complement(), `
			s0
				'a'     s1
				'b'     s0
			s1$
				'a'     s0
				'b'     s1
		`},
		{
			or(bsas, asbs.Complement()), `
			s0$
				'a'     s1
				'b'     s2
			s1
				'a'     s0
				'b'     s3
			s2$
				'a'     s3
				'b'     s0
			s3$
				'a'     s2
				'b'     s1
			`,
		},
		{
			and(bsas, asbs.Complement()), `
			s0
				'a'     s1
				'b'     s2
			s1
				'a'     s0
				'b'     s3
			s2$
				'a'     s3
				'b'     s0
			s3
				'a'     s2
				'b'     s1
			`,
		},
		{
			bsas.Exclude(asbs.Complement()), `
			s0$
				'a'     s1
				'b'     s2
			s1
				'a'     s0
				'b'     s3
			s2
				'a'     s3
				'b'     s0
			s3
				'a'     s2
				'b'     s1
			`,
		},
	} {
		expect(fmt.Sprintf("dump of test case %d", i), testcase.m.dump()).Equal(gspec.Unindent(testcase.s))
	}
}
