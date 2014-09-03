package scan

import (
	"testing"

	"github.com/hailiang/gspec"
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

func init() {
	//	fmt.Println(imaginaryLit)
}

func testMatch(t *testing.T, m Matcher, input string, expectedSize int, expectedMatched bool) {
	expect := gspec.Expect(t.FailNow, 1)
	size, matched := m.Match([]byte(input))
	expect("matched", matched).Equal(expectedMatched)
	expect("size of the match", size).Equal(expectedSize)
}
