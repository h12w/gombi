package scan

import (
	"regexp"
	"regexp/syntax"
)

type rx struct {
	*regexp.Regexp
}

type rxSyntax struct {
	*syntax.Regexp
}

func (r rxSyntax) isCharSet() bool {
	return r.Op == syntax.OpCharClass
}

func parse(p string) *syntax.Regexp {
	r, err := syntax.Parse(p, syntax.MatchNL|syntax.UnicodeGroups|syntax.PerlX)
	if err != nil {
		panic(err)
	}
	return r
}

var compile = regexp.MustCompile

func singleLiteralToCharClass(rx *syntax.Regexp) {
	if rx.Op == syntax.OpLiteral && len(rx.Rune) == 1 {
		char := rx.Rune[0]
		rx.Rune = []rune{char, char}
		rx.Op = syntax.OpCharClass
	}
}
