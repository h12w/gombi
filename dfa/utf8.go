package dfa

func BetweenRune(from, to rune) *Machine {
	if from > to {
		from, to = to, from
	}
	if from < 0 || to > 0x10FFFF {
		panic("invalid range for unicode point")
	}
	m := &Machine{}
	l := from
	if 0 <= l && l <= 0x7F {
		m = between1(l, rMin(0x7F, to))
		l = rMin(0x80, to)
	}
	if 0x80 <= l && l <= 0x7FF {
		m = Or(m, between2(l, rMin(0x7FF, to)))
		l = rMin(0x800, to)
	}
	if 0x800 <= l && l <= 0xFFFF {
		m = Or(m, between3(l, rMin(0xFFFF, to)))
		l = rMin(0x10000, to)
	}
	if 0x10000 <= l && l <= 0x10FFFF {
		m = Or(m, between4(l, rMin(0x10FFFF, to)))
	}
	return m.minimize()
}
func rMin(a, b rune) rune {
	if a < b {
		return a
	}
	return b
}

func between1(from, to rune) *Machine {
	return &Machine{states{
		stateBetween(byte(from), byte(to), 1),
		finalState(),
	}}
}

const (
	lead2 = 0xC0
	lead3 = 0xE0
	lead4 = 0xF0
	leadx = 0x80
	maskx = 0x3F
)

func between2(from, to rune) *Machine {
	f0, t0 := byte(from&maskx), byte(to&maskx)
	f1, t1 := byte(from>>6), byte(to>>6)
	if f1 == t1 {
		return &Machine{states{
			stateTo(f1|lead2, 1),
			stateBetween(f0|leadx, t0|leadx, 2),
			finalState(),
		}}
	}
	var m *Machine
	if f0 > 0 {
		m = &Machine{states{
			stateTo(f1|lead2, 1),
			stateBetween(f0|leadx, maskx|leadx, 2),
			finalState(),
		}}
		f1++
	}
	if t0 < maskx {
		m = Or(m, &Machine{states{
			stateTo(t1|lead2, 1),
			stateBetween(0x00|leadx, t0|leadx, 2),
			finalState(),
		}})
		t1--
	}
	if f1 <= t1 {
		m = Or(m, &Machine{states{
			stateBetween(f1|lead2, t1|lead2, 1),
			stateBetween(0x00|leadx, maskx|leadx, 2),
			finalState(),
		}})
	}
	return m
}

func between3(from, to rune) *Machine {
	f0, t0 := byte(from&maskx), byte(to&maskx)
	f1, t1 := byte((from>>6)&maskx), byte((to>>6)&maskx)
	f2, t2 := byte(from>>12), byte(to>>12)
	if f2 == t2 && f1 == t1 {
		return &Machine{states{
			stateTo(f2|lead3, 1),
			stateTo(f1|leadx, 2),
			stateBetween(f0|leadx, t0|leadx, 3),
			finalState(),
		}}
	}
	var m *Machine
	if f0 > 0 {
		m = Or(m, &Machine{states{
			stateTo(f2|lead3, 1),
			stateTo(f1|leadx, 2),
			stateBetween(f0|leadx, maskx|leadx, 3),
			finalState(),
		}})
		f1++
	}
	if t0 < maskx {
		m = Or(m, &Machine{states{
			stateTo(t2|lead3, 1),
			stateTo(t1|leadx, 2),
			stateBetween(0x00|leadx, t0|leadx, 3),
			finalState(),
		}})
		t1--
	}
	if f2 == t2 {
		m = Or(m, &Machine{states{
			stateTo(f2|lead3, 1),
			stateBetween(f1|leadx, t1|leadx, 2),
			stateBetween(0x00|leadx, maskx|leadx, 3),
			finalState(),
		}})
		return m
	}
	if f1 > 0 {
		m = Or(m, &Machine{states{
			stateTo(f2|lead3, 1),
			stateBetween(f1|leadx, maskx|leadx, 2),
			stateBetween(0x00|leadx, maskx|leadx, 3),
			finalState(),
		}})
		f2++
	}
	if t1 < maskx {
		m = Or(m, &Machine{states{
			stateTo(t2|lead3, 1),
			stateBetween(0x00|leadx, t1|leadx, 2),
			stateBetween(0x00|leadx, maskx|leadx, 3),
			finalState(),
		}})
		t2--
	}
	if f2 <= t2 {
		m = Or(m, &Machine{states{
			stateBetween(f2|lead3, t2|lead3, 1),
			stateBetween(0x00|leadx, maskx|leadx, 2),
			stateBetween(0x00|leadx, maskx|leadx, 3),
			finalState(),
		}})
	}
	return m
}

func between4(from, to rune) *Machine {
	return nil
}
