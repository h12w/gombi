package scan

const (
	RuneError    = '\uFFFD'
	MaxRune      = '\U0010FFFF'
	RuneSelf     = 0x80
	surrogateMin = 0xD800
	surrogateMax = 0xDFFF
	t1           = 0x00 // 0000 0000
	tx           = 0x80 // 1000 0000
	t2           = 0xC0 // 1100 0000
	t3           = 0xE0 // 1110 0000
	t4           = 0xF0 // 1111 0000
	t5           = 0xF8 // 1111 1000

	maskx = 0x3F // 0011 1111
	mask2 = 0x1F // 0001 1111
	mask3 = 0x0F // 0000 1111
	mask4 = 0x07 // 0000 0111

	rune1Max = 1<<7 - 1
	rune2Max = 1<<11 - 1
	rune3Max = 1<<16 - 1
)

type Matcher interface {
	Match(buf []byte) (int, bool)
	String() string
}

func (s *CharSet) Match(buf []byte) (int, bool) {
	if len(buf) == 0 {
		return 0, false
	}
	if s.a.contains(buf[0]) {
		return 1, true
	}
	r, size := decode(buf)
	if r != RuneError && s.u.contains(r) {
		return size, true
	}
	return 0, false
}
func decode(p []byte) (r rune, size int) {
	n := len(p)
	if n < 1 {
		return RuneError, 0
	}
	c0 := p[0]

	// 1-byte, 7-bit sequence?
	if c0 < tx {
		return rune(c0), 1
	}

	// unexpected
	// continuation byte?
	if c0 < t2 {
		return RuneError, 1
	}

	// need
	// first
	// continuation
	// byte
	if n < 2 {
		return RuneError, 1
	}
	c1 := p[1]
	if c1 < tx || t2 <= c1 {
		return RuneError, 1
	}

	// 2-byte,
	// 11-bit
	// sequence?
	if c0 < t3 {
		r = rune(c0&mask2)<<6 | rune(c1&maskx)
		if r <= rune1Max {
			return RuneError, 1
		}
		return r, 2
	}

	// need
	// second
	// continuation
	// byte
	if n < 3 {
		return RuneError, 1
	}
	c2 := p[2]
	if c2 < tx || t2 <= c2 {
		return RuneError, 1
	}

	// 3-byte,
	// 16-bit
	// sequence?
	if c0 < t4 {
		r = rune(c0&mask3)<<12 | rune(c1&maskx)<<6 | rune(c2&maskx)
		if r <= rune2Max {
			return RuneError, 1
		}
		if surrogateMin <= r && r <= surrogateMax {
			return RuneError, 1
		}
		return r, 3
	}

	// need
	// third
	// continuation
	// byte
	if n < 4 {
		return RuneError, 1
	}
	c3 := p[3]
	if c3 < tx || t2 <= c3 {
		return RuneError, 1
	}

	// 4-byte,
	// 21-bit
	// sequence?
	if c0 < t5 {
		r = rune(c0&mask4)<<18 | rune(c1&maskx)<<12 | rune(c2&maskx)<<6 | rune(c3&maskx)
		if r <= rune3Max || MaxRune < r {
			return RuneError, 1
		}
		return r, 4
	}

	// error
	return RuneError, 1
}
func (s *unicodeSet) contains(r rune) bool {
	for _, rg := range *s {
		if rg.s <= r && r <= rg.e {
			return true
		}
	}
	return false
}
func (r *asciiSet) contains(b byte) bool {
	if b < 64 {
		return r[0]&(1<<b) != 0
	} else if b < RuneSelf {
		return r[1]&(1<<(b-64)) != 0
	}
	return false
}

func (l *Literal) Match(buf []byte) (int, bool) {
	if len(buf) < len(l.a) {
		return 0, false
	}
	for i := range l.a {
		if l.a[i] != buf[i] {
			return 0, false
		}
	}
	return len(l.a), true
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
