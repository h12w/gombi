package scan

type TokenMatcher struct {
	Defs    []*IDMatcher
	EOF     int
	Illegal int
}
type IDMatcher struct {
	Matcher Matcher
	ID      int
}

func (m *TokenMatcher) Match(buf []byte) (id, size int) {
	if len(buf) == 0 {
		return m.EOF, 0
	}
	for _, d := range m.Defs {
		if size, ok := d.Matcher.Match(buf); ok {
			return d.ID, size
		}
	}
	return m.Illegal, 1 // advance 1 byte when illegal
}
