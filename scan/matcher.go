package scan

import (
	"fmt"
	"io"

	"h12.io/dfa"
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

func NewMatcher(eof, illegal int, mids []MID) *Matcher {
	m := or(mids)
	fast := m.ToFast()
	return &Matcher{
		EOF:     eof,
		Illegal: illegal,
		M:       m,
		fast:    fast}
}

func (m *Matcher) Init() *Matcher {
	m.fast = m.M.ToFast()
	return m
}

func (m *Matcher) Size() int {
	return m.fast.Size()
}

func (m *Matcher) Count() int {
	return m.fast.Count()
}

func (m *Matcher) WriteGo(w io.Writer, pac string) {
	fmt.Fprintln(w, "&scan.Matcher{")
	fmt.Fprintf(w, "EOF: %d,\n", m.EOF)
	fmt.Fprintf(w, "Illegal: %d,\n", m.Illegal)
	fmt.Fprint(w, "M: ")
	m.M.WriteGo(w, pac)
	fmt.Fprintln(w, "}")
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
	return dfa.Or(ms...)
}
