package re

import (
	"bytes"
	"unicode/utf8"
)

type Matcher interface {
	match(ctx *context, buf []byte) (int, bool)
	String() string
}

func (s *CharSet) match(_ *context, buf []byte) (int, bool) {
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

func (l *Literal) match(_ *context, buf []byte) (int, bool) {
	if len(buf) > len(l.a) {
		buf = buf[:len(l.a)]
	}
	if bytes.Equal(buf, l.a) {
		return len(buf), true
	}
	return 0, false
}

func (s *Sequence) match(ctx *context, buf []byte) (int, bool) {
	p := 0
	for _, m := range s.ms {
		if size, ok := m.match(ctx, buf[p:]); ok {
			p += size
		} else {
			return 0, false
		}
	}
	return p, true
}

func (s *Choice) match(ctx *context, buf []byte) (int, bool) {
	for _, m := range s.ms {
		if size, ok := m.match(ctx, buf); ok {
			return size, true
		}
	}
	return 0, false
}

func (r *Repetition) match(ctx *context, buf []byte) (int, bool) {
	p, n := 0, 0
	for {
		if r.sentinel != nil && n >= r.min {
			if size, ok := r.sentinel.match(ctx, buf[p:]); ok {
				p += size
				break
			}
		}
		if n == r.max {
			break
		}
		if size, ok := r.m.match(ctx, buf[p:]); ok {
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

func (c *Capturer) match(ctx *context, buf []byte) (int, bool) {
	size, ok := c.m.match(ctx, buf)
	if ok {
		ctx.capture(&token{id: c.id, value: buf[:size]})
	}
	return size, ok
}
