package scan

import "github.com/hailiang/dfa"

type Matcher struct {
	fast    *dfa.FastM
	eof     int
	illegal int
}
type MID struct {
	M  *dfa.M
	ID int
}

func NewMatcher(eof, illegal int, mids []MID) *Matcher {
	ms := make([]*dfa.M, len(mids))
	for i, mid := range mids {
		mid.M.As(mid.ID)
		ms[i] = mid.M
	}
	return &Matcher{
		eof:     eof,
		illegal: illegal,
		fast:    dfa.Or(ms...).Minimize().ToFast()}
}

func (m *Matcher) Size() int {
	return m.fast.Size()
}

//func (m *Matcher) SaveSVG(file string) error {
//	return m.m.SaveSVG(file)
//}
//
//func (m *Matcher) SaveDot(file string) error {
//	return m.m.SaveDot(file)
//}
