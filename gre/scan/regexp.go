package scan

import (
	"io"
	"regexp"
	"regexp/syntax"
	"unicode"
)

const batchSize = 1024

type Matcher struct {
	*regexp.Regexp
	ids []int
}

func NewMatcher(es ...Expr) *Matcher {
	pat := Con(Pat(`\A`), exprs(es).capture(true))
	return &Matcher{regexp.MustCompile(pat.String()), nil}
}

func NewMapMatcher(mm MM) *Matcher {
	es := make([]Expr, len(mm))
	ids := make([]int, len(mm))
	for i := range mm {
		es[i], ids[i] = mm[i].Expr, mm[i].ID
	}
	m := NewMatcher(es...)
	m.ids = ids
	return m
}

type MM []struct {
	Expr Expr
	ID   int
}

func (m *Matcher) matchBytes(buf []byte) (id, size int) {
	return m.result(m.FindSubmatchIndex(buf), 0)
}

func (m *Matcher) scanBatch(buf []byte, start int) []Token {
	rs := m.FindAllSubmatchIndex(buf, batchSize)
	toks := make([]Token, len(rs))
	p := 0
	for i, match := range rs {
		id, size := m.result(match, p)
		toks[i].ID = id
		toks[i].Value = buf[p : p+size]
		toks[i].Pos = start + p
		p += size
	}
	return toks
}

func (m *Matcher) matchReader(r io.RuneReader) (id, size int) {
	return m.result(m.FindReaderSubmatchIndex(r), 0)
}

func (m *Matcher) result(match []int, start int) (id, size int) {
	if match != nil {
		for i := 2; i < len(match)-1; i += 2 {
			if match[i] == start {
				size = match[i+1] - match[i] // m[i] must be 0
				id = i / 2
				if m.ids != nil {
					id = m.ids[id-1]
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
