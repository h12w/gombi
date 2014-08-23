package re

import (
	"testing"

	"github.com/hailiang/gspec"
)

var (
	c          = Char
	b          = Between
	merge      = Merge
	s          = Str
	con        = Con
	or         = Or
	zeroOrOne  = ZeroOrOne
	zeroOrMore = ZeroOrMore
	oneOrMore  = OneOrMore
	repeat     = Repeat
	cap        = Capture

	illegal     = c("\x00")
	any         = illegal.Negate()
	newline     = c("\n")
	unicodeChar = any.Exclude(newline)
	//	unicodeLetter = c(`\p{L}`)
	//	unicodeDigit  = c(`\p{Nd}`)
	unicodeLetter = merge(b('A', 'Z'), b('a', 'z'), c(`۰۱۸६४ŝ`))
	unicodeDigit  = merge(decimalDigit, c(`９８７６`))
	letter        = merge(unicodeLetter, c(`_`))
	decimalDigit  = b('0', '9')
	octalDigit    = b('0', '7')
	hexDigit      = merge(b('0', '9'), b('A', 'F'), b('a', 'f'))

	empty = s(``)

	whitespaces          = oneOrMore(c(" \t\r"))
	lineComment          = con(s(`//`), zeroOrMore(unicodeChar), or(newline, empty))
	generalCommentSL     = con(s(`/*`), zeroOrMore(any.Exclude(newline)).EndWith(s(`*/`)))
	generalCommentML     = con(s(`/*`), zeroOrMore(any).EndWith(s(`*/`)))
	identifier           = con(letter, zeroOrMore(or(letter, unicodeDigit)))
	intLit               = or(hexLit, decimalLit, octalLit)
	decimalLit           = con(b('1', '9'), zeroOrMore(decimalDigit))
	octalLit             = con(s(`0`), zeroOrMore(octalDigit))
	hexLit               = con(s(`0`), c("xX"), oneOrMore(hexDigit))
	floatLit             = or(floatLit1, floatLit2, floatLit3)
	floatLit1            = con(decimals, s(`.`), zeroOrOne(decimals), zeroOrOne(exponent))
	floatLit2            = con(decimals, exponent)
	floatLit3            = con(s(`.`), decimals, zeroOrOne(exponent))
	decimals             = oneOrMore(decimalDigit)
	exponent             = con(c("eE"), zeroOrOne(c("+-")), decimals)
	imaginaryLit         = con(or(decimals, floatLit), s(`i`))
	runeLit              = con(c("'"), or(unicodeValue, byteValue), c("'"))
	unicodeValue         = or(unicodeChar, littleUValue, bigUValue, escapedChar)
	unicodeStrValue      = or(unicodeChar.Exclude(c(`"`)), littleUValue, bigUValue, escapedChar)
	byteValue            = or(octalByteValue, hexByteValue)
	octalByteValue       = con(s(`\`), repeat(octalDigit, 3))
	hexByteValue         = con(s(`\x`), repeat(hexDigit, 2))
	littleUValue         = con(s(`\u`), repeat(hexDigit, 4))
	bigUValue            = con(s(`\U`), repeat(hexDigit, 8))
	escapedChar          = con(s(`\`), c(`abfnrtv\'"`))
	rawStringLit         = con(s("`"), zeroOrMore(or(unicodeChar.Exclude(c("`")), newline)), s("`"))
	interpretedStringLit = con(s(`"`), zeroOrMore(or(unicodeStrValue, byteValue)), s(`"`))
)

func TestMatch(t *testing.T) {
	testMatch(t, c("ac"), "ac", 1, true)
	testMatch(t, c("ac"), "ca", 1, true)
	testMatch(t, c("ac"), "b", 0, false)
	l := Str("xy")
	testMatch(t, l, "xy", 2, true)
	testMatch(t, l, "x", 0, false)
	testMatch(t, l, "y", 0, false)
	testMatch(t, con(c("ac"), l), "axy", 3, true)
	testMatch(t, con(c("ac"), l), "cxy", 3, true)
	testMatch(t, con(c("ac"), l), "a", 0, false)
	cho := Or(c("ac"), l)
	testMatch(t, cho, "a", 1, true)
	testMatch(t, cho, "xy", 2, true)
	testMatch(t, cho, "x", 0, false)
	testMatch(t, zeroOrOne(s(`a`)), "", 0, true)
	testMatch(t, zeroOrOne(s(`a`)), "a", 1, true)
	testMatch(t, zeroOrOne(s(`a`)), "aa", 1, true)
	testMatch(t, zeroOrMore(s(`a`)), "", 0, true)
	testMatch(t, zeroOrMore(s(`a`)), "a", 1, true)
	testMatch(t, zeroOrMore(s(`a`)), "aa", 2, true)
	testMatch(t, oneOrMore(s(`a`)), "", 0, false)
	testMatch(t, oneOrMore(s(`a`)), "a", 1, true)
	testMatch(t, oneOrMore(s(`a`)), "aa", 2, true)
	testMatch(t, repeat(s(`a`), 0), "", 0, true)
	testMatch(t, repeat(s(`a`), 0), "a", 0, true)
	testMatch(t, repeat(s(`a`), 1), "", 0, false)
	testMatch(t, repeat(s(`a`), 1), "a", 1, true)
	testMatch(t, repeat(s(`a`), 1), "aa", 1, true)
	testMatch(t, repeat(s(`a`), 2), "a", 0, false)
	testMatch(t, repeat(s(`a`), 2), "aa", 2, true)
	testMatch(t, repeat(s(`a`), 2), "aaa", 2, true)
	testMatch(t, zeroOrMore(c(`a*/`)), "aa*/aa*/", 8, true)
	testMatch(t, zeroOrMore(c(`a*/`)).EndWith(s(`*/`)), "aa*/aa*/", 4, true)
}

func TestCapture(t *testing.T) {
	expect := gspec.Expect(t.FailNow)
	m := cap(1, s(`a`))
	ctx := &context{}
	m.match(ctx, []byte("ab"))
	expect(ctx.tokens).Equal([]*token{&token{id: 1, value: []byte("a")}})
}

func init() {
	//	fmt.Println(imaginaryLit)
}

func testMatch(t *testing.T, m Matcher, input string, expectedSize int, expectedMatched bool) {
	expect := gspec.Expect(t.FailNow, 1)
	size, matched := m.match(nil, []byte(input))
	expect("matched", matched).Equal(expectedMatched)
	expect("size of the match", size).Equal(expectedSize)
}
