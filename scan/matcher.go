package scan

import "github.com/hailiang/dfa"

type Matcher struct {
	m       *dfa.M
	fast    *dfa.FastM
	eof     int
	illegal int
}
type MID struct {
	M  *dfa.M
	ID int
}

func NewMatcher(eof, illegal int, mids []MID) *Matcher {
	m := or(mids)
	fast := m.ToFast()
	return &Matcher{
		eof:     eof,
		illegal: illegal,
		m:       m,
		fast:    fast}
}

func (m *Matcher) Size() int {
	return m.fast.Size()
}

func (m *Matcher) Count() int {
	return m.fast.Count()
}

func (m *Matcher) SaveSVG(file string) error {
	return m.m.SaveSVG(file)
}

func (m *Matcher) SaveDot(file string) error {
	return m.m.SaveDot(file)
}

func (m *Matcher) String() string {
	return m.m.String()
}

func or(mids []MID) *dfa.M {
	ms := make([]*dfa.M, len(mids))
	for i, mid := range mids {
		mid.M.As(mid.ID)
		ms[i] = mid.M
	}
	return dfa.Or(ms...).Minimize()
}

func GenGo(mids []MID, file, pac string) error {
	return or(mids).SaveGo(file, pac)
}

//func (m *Matcher) SaveSVG(file string) error {
//	return m.m.SaveSVG(file)
//}
//
//func (m *Matcher) SaveDot(file string) error {
//	return m.m.SaveDot(file)
//}
