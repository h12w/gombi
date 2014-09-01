package dfa

const (
	leadx = 0x80
	maskx = 0x3F
)

func (u *u8s) between(lo, hi rune) *Machine {
	if lo > hi {
		lo, hi = hi, lo
	}
	if lo < 0 || hi > 0x10ffff {
		panic("invalid range for unicode point")
	}
	l := lo
	if 0 <= l && l <= 0x7F {
		u.betweenASCII(l, rMin(0x7F, hi))
		l = rMin(0x80, hi)
	}
	u.betweenRune1(l, hi)
	return u.m() //.minimize()
}

func (u *u8s) betweenRune1(lo, hi rune) {
	if lo > hi {
		return
	}
	limit := []rune{
		0x7FF,
		0xFFFF,
	}
	for _, l := range limit {
		if lo <= l && hi > l {
			u.betweenRune1(lo, l)
			u.betweenRune1(l+1, hi)
			return
		}
	}
	mask := []rune{
		0x0003f,
		0x00fff,
		0x3ffff,
	}
	for _, m := range mask {
		if lo&^m != hi&^m {
			if lo&m != 0 {
				u.betweenRune1(lo, lo|m)
				u.betweenRune1((lo|m)+1, hi)
				return
			}
			if hi&m != m {
				u.betweenRune1(lo, (hi&^m)-1)
				u.betweenRune1(hi&^m, hi)
				return
			}
		}
	}
	u.or(u8m{lo, hi})
	for i := 0; lo != 0; i++ {
		b := byte(lo&maskx | leadx)
		lo >>= 6
		if lo == 0 {
			b |= lead[i+1]
		}
	}
	for i := 0; hi != 0; i++ {
		b := byte(hi&maskx | leadx)
		hi >>= 6
		if hi == 0 {
			b |= lead[i+1]
		}
	}
}

/*
func (u *u8s) betweenRune(lo, hi rune) {
	l := lo
	if 0x80 <= l && l <= 0x7FF {
		u.between2(l, rMin(0x7FF, hi))
		l = rMin(0x800, hi)
	}
	if 0x800 <= l && l <= 0xFFFF {
		u.between3(l, rMin(0xFFFF, hi))
		l = rMin(0x10000, hi)
	}
	if 0x10000 <= l && l <= 0x10FFFF {
		u.between4(l, rMin(0x10FFFF, hi))
	}
}
*/
func rMin(a, b rune) rune {
	if a < b {
		return a
	}
	return b
}

func (u *u8s) betweenASCII(lo, hi rune) {
	u.or(u8m{lo, hi})
}

/*
func (u *u8s) between2(lo, hi rune) {
	f0, t0 := b6(lo, 0), b6(hi, 0)
	f1, t1 := b6(lo, 1), b6(hi, 1)
	if f1 == t1 {
		u.or(u8{
			{f1, f1},
			{f0, t0},
		})
		return
	}
	if f0 > 0 {
		u.or(u8{
			{f1, f1},
			{f0, maskx},
		})
		f1++
	}
	if t0 < maskx {
		u.or(u8{
			{t1, t1},
			{0, t0},
		})
		t1--
	}
	if f1 <= t1 {
		u.or(u8{
			{f1, t1},
			{0, maskx},
		})
	}
}

func (u *u8s) between3(lo, hi rune) {
	f0, t0 := b6(lo, 0), b6(hi, 0)
	f1, t1 := b6(lo, 1), b6(hi, 1)
	f2, t2 := b6(lo, 2), b6(hi, 2)
	if f2 == t2 && f1 == t1 {
		u.or(u8{
			{f2, f2},
			{f1, f1},
			{f0, t0},
		})
		return
	}
	if f0 > 0 {
		u.or(u8{
			{f2, f2},
			{f1, f1},
			{f0, maskx},
		})
		f1++
	}
	if t0 < maskx {
		u.or(u8{
			{t2, t2},
			{t1, t1},
			{0, t0},
		})
		t1--
	}
	if f2 == t2 {
		u.or(u8{
			{f2, f2},
			{f1, t1},
			{0, maskx},
		})
		return
	}
	if f1 > 0 {
		u.or(u8{
			{f2, f2},
			{f1, maskx},
			{0, maskx},
		})
		f2++
	}
	if t1 < maskx {
		u.or(u8{
			{t2, t2},
			{0, t1},
			{0, maskx},
		})
		t2--
	}
	if f2 <= t2 {
		u.or(u8{
			{f2, t2},
			{0, maskx},
			{0, maskx},
		})
	}
}

func (u *u8s) between4(lo, hi rune) {
}
*/

// u8s represents the union of multiple u8's and is used hi construct an UTF8
// DFA by collecting u8's.
type u8s struct {
	a []u8m
}

func (s *u8s) or(u u8m) {
	s.a = append(s.a, u)
}

func (s *u8s) m() *Machine {
	var m *Machine
	for i := range s.a {
		m = or2(m, s.a[i].m())
	}
	return m
}

// u8 represents a single UTF8 range
//type u8 [][2]byte
//
var lead = []byte{
	0x00,
	0x00, // 1 byte
	0xC0, // 2 bytes
	0xE0, // 3 bytes
	0xF0, // 4 bytes
}

//
//func (u u8) m() *Machine {
//	n := len(u)
//	if n == 0 || n > 4 {
//		panic("invalid UTF8 specification")
//	}
//	m := &Machine{make(states, n+1)}
//	m.states[0] = stateBetween(u[0][0]|lead[n], u[0][1]|lead[n], 1)
//	for i := 1; i < n; i++ {
//		m.states[i] = stateBetween(u[i][0]|leadx, u[i][1]|leadx, i+1)
//	}
//	m.states[len(m.states)-1] = finalState()
//	return m
//}

// b6 right shift r by 6*n bits and gets the lower 6 bits.
func b6(r rune, n int) byte {
	for i := 0; i < n; i++ {
		r >>= 6
	}
	return byte(r & maskx)
}

type u8m struct {
	lo, hi rune
}

func (u *u8m) m() *Machine {
	lo, hi := u.lo, u.hi
	if hi < 0x80 {
		return &Machine{states{
			stateBetween(byte(lo), byte(hi), 1),
			finalState(),
		}}
	}
	bs := make([][2]byte, 0, 4)
	n := 1
	if 0x0080 <= hi && hi <= 0x07ff {
		n = 2
	} else if 0x0800 <= hi && hi <= 0xffff {
		n = 3
	} else if hi <= 0x10ffff {
		n = 4
	}
	for i := 0; i < n; i++ {
		bs = append(bs, [2]byte{byte(lo & maskx), byte(hi & maskx)})
		lo >>= 6
		hi >>= 6
	}
	//fmt.Printf("%.2x ", bs[n-1][0]|lead[n])
	//fmt.Printf("%.2x ", bs[n-1][1]|lead[n])
	//fmt.Println()
	m := &Machine{make(states, n+1)}
	m.states[0] = stateBetween(bs[n-1][0]|lead[n], bs[n-1][1]|lead[n], 1)
	for i := 1; i < n; i++ {
		m.states[i] = stateBetween(bs[n-i-1][0]|leadx, bs[n-i-1][1]|leadx, i+1)
	}
	m.states[len(m.states)-1] = finalState()
	return m
}
