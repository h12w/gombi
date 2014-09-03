package scan

import (
	"sort"
	"unicode"
)

//var anyChar = Between(0, unicode.MaxRune)
const maxInt = 1<<31 - 1

type (
	CharSet struct {
		a asciiSet
		u unicodeSet
	}
	runeRange struct { // rune range
		s, e rune
	}
	asciiSet   [2]uint64
	unicodeSet []runeRange

	Literal struct {
		a []byte
	}

	Sequence struct {
		ms []Matcher
	}

	Choice struct {
		ms []Matcher
	}

	Repetition struct {
		m        Matcher
		sentinel Matcher
		min      int
		max      int
	}

	runes []rune
)

func Char(s string) *CharSet {
	rs := runes(s)
	sort.Sort(rs)
	return &CharSet{rs.asciiRunes().toAsciiSet(), rs.unicodeRunes().toUnicodeSet()}
}
func (rs runes) asciiRunes() runes {
	for i, r := range rs {
		if r >= RuneSelf {
			return rs[:i]
		}
	}
	return rs
}
func (rs runes) unicodeRunes() runes {
	for i, r := range rs {
		if r >= RuneSelf {
			return rs[i:]
		}
	}
	return nil
}
func (rs runes) toAsciiSet() (a asciiSet) {
	for _, r := range rs {
		a.add(byte(r))
	}
	return
}
func (rs runes) toUnicodeSet() (rr unicodeSet) {
	if len(rs) == 0 {
		return nil
	}
	rr = append(rr, runeRange{s: rs[0]})
	cur := rs[0]
	for i := 1; i < len(rs); i++ {
		if rs[i] > cur+1 {
			rr[len(rr)-1].e = cur
			rr = append(rr, runeRange{s: rs[i]})
		}
		cur = rs[i]
	}
	rr[len(rr)-1].e = cur
	return
}
func (rs runes) Len() int           { return len(rs) }
func (rs runes) Less(i, j int) bool { return rs[i] < rs[j] }
func (rs runes) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

func Between(s, e byte) *CharSet {
	if s > e {
		s, e = e, s
	}
	var a asciiSet
	for b := s; b <= e; b++ {
		a.add(b)
	}
	return &CharSet{a: a}
}

func Str(s string) *Literal {
	return &Literal{[]byte(s)}
}

func Con(ms ...Matcher) *Sequence {
	return &Sequence{ms}
}

func Or(ms ...Matcher) *Choice {
	return &Choice{ms}
}

func Merge(cs ...*CharSet) *CharSet {
	var a asciiSet
	rrs := make(unicodeSet, 0)
	for _, c := range cs {
		a.merge(&c.a)
		rrs = append(rrs, unicodeSet(c.u)...)
	}
	sort.Sort(rrs)
	return &CharSet{a: a, u: rrs.simplify()}
}
func (rs unicodeSet) Len() int           { return len(rs) }
func (rs unicodeSet) Less(i, j int) bool { return rs[i].s < rs[j].s }
func (rs unicodeSet) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }
func (rs unicodeSet) simplify() unicodeSet {
	if len(rs) < 2 {
		return rs
	}
	cur := 0
	for i := 1; i < len(rs); i++ {
		if rs[cur].e >= rs[i].s {
			rs[cur].e = rs[i].e
		} else {
			cur++
		}
	}
	return rs[:cur+1]
}

func (s *CharSet) Negate() *CharSet {
	return &CharSet{s.a.negate(), unicodeSet(s.u).negate()}
}
func (rs unicodeSet) negate() (neg unicodeSet) {
	min := rune(RuneSelf)
	for _, r := range rs {
		s, e := r.s, r.e
		if min < s {
			neg = append(neg, runeRange{min, s - 1})
		}
		min = e + 1
	}
	if min <= unicode.MaxRune {
		neg = append(neg, runeRange{min, unicode.MaxRune})
	}
	return neg
}

func (s *CharSet) Exclude(cs ...*CharSet) *CharSet {
	return Merge(s.Negate(), Merge(cs...)).Negate()
}

func ZeroOrMore(m Matcher) *Repetition {
	return &Repetition{m, nil, 0, maxInt}
}

func OneOrMore(m Matcher) *Repetition {
	return &Repetition{m, nil, 1, maxInt}
}

func ZeroOrOne(m Matcher) *Repetition {
	return &Repetition{m, nil, 0, 1}
}

func Repeat(m Matcher, n int) *Repetition {
	return &Repetition{m, nil, n, n}
}

func (r *Repetition) EndWith(sentinel Matcher) *Repetition {
	r.sentinel = sentinel
	return r
}

func (r *asciiSet) add(b byte) {
	if b < 64 {
		r[0] |= (1 << b)
	}
	r[1] |= (1 << (b - 64))
}

func (r *asciiSet) merge(o *asciiSet) {
	(*r)[0] |= (*o)[0]
	(*r)[1] |= (*o)[1]
}

func (r *asciiSet) negate() asciiSet {
	neg := *r
	neg[0] = ^neg[0]
	neg[1] = ^neg[1]
	return neg
}
