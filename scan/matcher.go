package scan

import (
	"encoding/gob"
	"os"

	"github.com/hailiang/dfa"
)

type Matcher struct {
	*dfa.M
	EOF     int
	Illegal int

	fast *dfa.FastM
}
type MID struct {
	M  interface{}
	ID int
}

func LoadMatcher(file string) (*Matcher, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var m Matcher
	err = gob.NewDecoder(f).Decode(&m)
	if err != nil {
		return nil, err
	}
	m.fast = m.M.ToFast()
	return &m, nil
}

func (m *Matcher) SaveCache(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(m)
}

func NewMatcher(eof, illegal int, mids []MID) *Matcher {
	m := or(mids)
	fast := m.ToFast()
	return &Matcher{
		EOF:     eof,
		Illegal: illegal,
		M:       m,
		fast:    fast}
}

func (m *Matcher) Size() int {
	return m.fast.Size()
}

func (m *Matcher) Count() int {
	return m.fast.Count()
}

func or(mids []MID) *dfa.M {
	ms := make([]interface{}, len(mids))
	for i, mid := range mids {
		var m *dfa.M
		switch o := mid.M.(type) {
		case *dfa.M:
			m = o
		case string:
			m = dfa.Str(o)
		default:
			panic("member M of MID should be type of either string or *M")
		}
		ms[i] = m.As(mid.ID)
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
