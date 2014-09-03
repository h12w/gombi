package scan

import (
	"fmt"

	"github.com/hailiang/dfa"
)

type Matcher struct {
	fast    *dfa.FastM
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
		fast:    m.ToFast()}
}

//func (m *Matcher) Match(buf []byte) (id, size int) {
//	if len(buf) == 0 {
//		id, size = m.eof, 0
//	} else if si, label, ok := m.m.Match(buf); ok {
//		id, size = label, si
//	} else {
//		id, size = m.illegal, 1 // advance 1 byte when illegal
//	}
//	return
//}

//func (m *Matcher) SaveSVG(file string) error {
//	return m.m.SaveSVG(file)
//}
//
//func (m *Matcher) SaveDot(file string) error {
//	return m.m.SaveDot(file)
//}
