package lex

import (
	"bytes"
	"fmt"
	"regexp/syntax"
	"unicode"
)

type regexp struct {
	*syntax.Regexp
}

type Pattern struct {
	regexp
}

func (p Pattern) Negate() Pattern {
	switch p.Op {
	case syntax.OpLiteral:
		if len(p.Rune) == 1 {
			return Char(`^` + p.String())
		}
	case syntax.OpCharClass:
		neg := p.Regexp
		neg.Rune = negateRune(neg.Rune)
		neg.Rune0[0] = neg.Rune[0]
		neg.Rune0[1] = neg.Rune[1]
		return Pattern{regexp{neg}}
	}
	panic(fmt.Errorf("Pattern %#v is not a character range.", p.Regexp))
}

func (p Pattern) Exclude(s Pattern) Pattern {
	return Or(p.Negate(), s).Negate()
}

func negateRune(rs []rune) []rune {
	if len(rs) == 0 {
		panic(fmt.Errorf("unexpected empty []rune"))
	}
	neg := []rune{}
	min := rMin(0, rs[0])
	for i := 0; i < len(rs)-1; i += 2 {
		l, r := rs[i], rs[i+1]
		if min < l {
			neg = append(neg, min)
			neg = append(neg, l-1)
		}
		min = r + 1
	}
	if min <= unicode.MaxRune {
		neg = append(neg, min)
		neg = append(neg, unicode.MaxRune)
	}
	return neg
}

func rMin(a, b rune) rune {
	if a < b {
		return a
	}
	return b
}

func Char(p string) Pattern {
	if p == "." {
		return Pat(p)
	}
	pat := Pat("[" + p + "]")
	if pat.Op == syntax.OpLiteral && len(pat.Rune) == 1 {
		char := pat.Rune[0]
		pat.Rune = []rune{char, char}
		pat.Op = syntax.OpCharClass
	}
	return pat
}

func Pat(p string) Pattern {
	return Pattern{regexp{parse(p)}}
}

func Or(ps ...Pattern) Pattern {
	var buf bytes.Buffer
	for i, p := range ps {
		if i > 0 {
			buf.WriteByte('|')
		}
		buf.WriteString("(?:")
		buf.WriteString(p.String())
		buf.WriteByte(')')
	}
	return Pat(buf.String())
}

func parse(p string) *syntax.Regexp {
	r, err := syntax.Parse(p, syntax.MatchNL|syntax.UnicodeGroups|syntax.PerlX)
	if err != nil {
		panic(err)
	}
	return r
}
