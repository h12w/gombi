package scan

import (
	"io"
	"regexp"
	"regexp/syntax"
)

type Matcher struct {
	*regexp.Regexp
	ids []int
}

func NewMatcher(es ...Expr) *Matcher {
	pat := Con(Pat(`\A`), exprs(es).capture(true))
	return &Matcher{regexp.MustCompile(pat.String()), nil}
}

func (m *Matcher) Map(ids ...int) *Matcher {
	m.ids = ids
	return m
}

func (m *Matcher) matchBytes(buf []byte) (id, size int) {
	return m.result(m.FindSubmatchIndex(buf))
}

func (m *Matcher) matchReader(r io.RuneReader) (id, size int) {
	return m.result(m.FindReaderSubmatchIndex(r))
}

func (m *Matcher) result(match []int) (id, size int) {
	if match != nil {
		for i := 2; i < len(match)-1; i += 2 {
			if match[i] != -1 {
				size = match[i+1] // m[i] must be 0
				id = i/2 - 1
				if m.ids != nil {
					id = m.ids[id]
				}
				return id, size
			}
		}
	}
	return -1, 0
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

func singleLiteralToCharClass(rx *syntax.Regexp) {
	if rx.Op == syntax.OpLiteral && len(rx.Rune) == 1 {
		char := rx.Rune[0]
		rx.Rune = []rune{char, char}
		rx.Op = syntax.OpCharClass
	}
}
