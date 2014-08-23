package scan

import (
	"bytes"
	"unicode/utf8"
)

type Matcher interface {
	Match(buf []byte) (int, bool)
	String() string
}

func (s *CharSet) Match(buf []byte) (int, bool) {
	r, size := utf8.DecodeRune(buf)
	if r == utf8.RuneError {
		if len(buf) > 0 && s.contains(rune(buf[0])) {
			return 1, true
		}
	} else if s.contains(r) {
		return size, true
	}
	return 0, false
}
func (s *CharSet) contains(r rune) bool {
	for _, rg := range s.ranges {
		if rg.s <= r && r <= rg.e {
			return true
		}
	}
	return false
}

func (l *Literal) Match(buf []byte) (int, bool) {
	if len(buf) > len(l.a) {
		buf = buf[:len(l.a)]
	}
	if bytes.Equal(buf, l.a) {
		return len(buf), true
	}
	return 0, false
}

func (s *Sequence) Match(buf []byte) (int, bool) {
	p := 0
	for _, m := range s.ms {
		if size, ok := m.Match(buf[p:]); ok {
			p += size
		} else {
			return 0, false
		}
	}
	return p, true
}

func (s *Choice) Match(buf []byte) (int, bool) {
	for _, m := range s.ms {
		if size, ok := m.Match(buf); ok {
			return size, true
		}
	}
	return 0, false
}

func (r *Repetition) Match(buf []byte) (int, bool) {
	p, n := 0, 0
	for {
		if r.sentinel != nil && n >= r.min {
			if size, ok := r.sentinel.Match(buf[p:]); ok {
				p += size
				break
			}
		}
		if n == r.max {
			break
		}
		if size, ok := r.m.Match(buf[p:]); ok {
			p += size
			n++
		} else {
			break
		}
	}
	if r.min <= n && n <= r.max {
		return p, true
	}
	return 0, false
}
