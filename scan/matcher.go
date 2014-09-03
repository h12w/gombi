package scan

import (
	"fmt"

	"github.com/hailiang/dfa"
)

type Matcher struct {
	m       *dfa.Machine
	eof     int
	illegal int
}
type MID struct {
	M  *dfa.Machine
	ID int
}

func NewMatcher(eof, illegal int, mids []MID) *Matcher {
	ms := make([]*dfa.Machine, len(mids))
	for i, mid := range mids {
		mid.M.As(mid.ID)
		ms[i] = mid.M
	}
	m := dfa.Or(ms...)
	fmt.Println(m.Count())
	m = m.Minimize()
	fmt.Println(m.Count())
	return &Matcher{
		eof:     eof,
		illegal: illegal,
		m:       m}
}

func (m *Matcher) Match(buf []byte) (id, size int) {
	if len(buf) == 0 {
		return m.eof, 0
	}
	if size, label, ok := m.m.Match(buf); ok {
		return label, size
	}
	return m.illegal, 1 // advance 1 byte when illegal
}

func (m *Matcher) SaveSVG(file string) error {
	return m.m.SaveSVG(file)
}

func (m *Matcher) SaveDot(file string) error {
	return m.m.SaveDot(file)
}
