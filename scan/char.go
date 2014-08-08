package scan

import (
	"fmt"
	"unicode"
)

type CharSet struct {
	Pattern
}

func newCharSet(p Pattern) CharSet {
	c := CharSet{p}
	singleLiteralToCharClass(c.Regexp)
	if !c.isCharSet() {
		panic(fmt.Errorf("Pattern %s is not a character set.", c.String()))
	}
	return c
}

func Char(p string) CharSet {
	return newCharSet(Pat("[" + p + "]"))
}

func (p CharSet) Negate() CharSet {
	neg := newCharSet(Pat(p.String()))
	neg.Rune = negateRune(neg.Rune)
	neg.Rune0[0] = neg.Rune[0]
	neg.Rune0[1] = neg.Rune[1]
	return neg
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

func (p CharSet) Exclude(cs ...CharSet) CharSet {
	es := make(exprs, len(cs))
	for i := range cs {
		es[i] = cs[i]
	}
	subset := Merge(es...)
	return Merge(p.Negate(), subset).Negate()
}

func Merge(es ...Expr) CharSet {
	return newCharSet(exprs(es).or(false))
}
