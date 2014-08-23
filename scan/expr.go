package scan

import (
	"sort"
	"unicode"
)

//var anyChar = Between(0, unicode.MaxRune)
const maxInt = 1<<31 - 1

type (
	CharSet struct {
		ranges []runeRange
	}
	runeRange struct { // rune range
		s, e rune
	}

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

	runes      []rune
	runeRanges []runeRange
)

func Char(s string) *CharSet {
	return &CharSet{runes(s).toRanges()}
}
func (rs runes) toRanges() (rr []runeRange) {
	if len(rs) == 0 {
		return nil
	}
	sort.Sort(rs)
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

func Between(s, e rune) *CharSet {
	if s > e {
		s, e = e, s
	}
	return &CharSet{[]runeRange{{s, e}}}
}

//func Set(name string) *CharSet {
//}

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
	rrs := make(runeRanges, 0)
	for _, c := range cs {
		rrs = append(rrs, runeRanges(c.ranges)...)
	}
	sort.Sort(rrs)
	return &CharSet{rrs.simplify()}
}
func (rs runeRanges) Len() int           { return len(rs) }
func (rs runeRanges) Less(i, j int) bool { return rs[i].s < rs[j].s }
func (rs runeRanges) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }
func (rs runeRanges) simplify() runeRanges {
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
	return &CharSet{runeRanges(s.ranges).negate()}
}
func (rs runeRanges) negate() (neg []runeRange) {
	min := rune(0)
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
