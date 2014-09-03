package scan

import (
	"regexp"
	"regexp/syntax"
	"unicode"
)

/*
Things that have been tried but is worse (from slow to even slower):
1. Use a large expression to capture each alterantive submatch (FindSubmatchIndex) like:
   \A(?:(token1)|(token2)\(token3))

2. Find consecutive tokens all at once (FindAllSubmatchIndex):
	(token1)|(token2)\(token3)
*/

type Matcher struct {
	Defs    []*IDMatcher
	EOF     int
	Illegal int
}

func NewMatcher(es ...Expr) *Matcher {
	defs := make([]*IDMatcher, len(es))
	for i := range es {
		defs[i] = &IDMatcher{
			regexp.MustCompile(Con(Pat(`\A`), es[i]).String()),
			i + 1}
	}
	return &Matcher{Defs: defs}
}

func NewMapMatcher(mm MM) *Matcher {
	defs := make([]*IDMatcher, len(mm))
	for i := range mm {
		defs[i] = &IDMatcher{
			regexp.MustCompile(Con(Pat(`\A`), mm[i].Expr).String()),
			mm[i].ID}
	}
	return &Matcher{Defs: defs}
}

type MM []struct {
	Expr Expr
	ID   int
}
type IDMatcher struct {
	*regexp.Regexp
	ID int
}

func (m *Matcher) matchBytes(buf []byte) (id, size int) {
	if len(buf) == 0 {
		return m.EOF, 0
	}
	for _, d := range m.Defs {
		if loc := d.Regexp.FindIndex(buf); loc != nil {
			return d.ID, loc[1]
		}
	}
	return m.Illegal, 1
}

type rxSyntax struct {
	*syntax.Regexp
}

func (r rxSyntax) isCharSet() bool {
	return r.Op == syntax.OpCharClass
}

func parsePat(p string) *syntax.Regexp {
	r, err := syntax.Parse(p, syntax.ClassNL|syntax.DotNL|syntax.MatchNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		panic(err)
	}
	return r
}

func parseStr(p string) *syntax.Regexp {
	r, err := syntax.Parse(p, syntax.Literal)
	if err != nil {
		panic(err)
	}
	return r
}

func singleLiteralToCharClass(rx *syntax.Regexp) {
	if rx.Op == syntax.OpLiteral && len(rx.Rune) == 1 {
		char := rx.Rune[0]
		if rx.Flags&syntax.FoldCase != 0 && unicode.ToLower(char) != unicode.ToUpper(char) {
			l, h := unicode.ToLower(char), unicode.ToUpper(char)
			rx.Rune = []rune{h, h, l, l}
			rx.Rune0 = [...]rune{h, h}
		} else {
			rx.Rune = []rune{char, char}
			rx.Rune0 = [...]rune{char, char}
		}
		rx.Op = syntax.OpCharClass
	}
}
