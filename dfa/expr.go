package dfa

import "unicode/utf8"

func Str(s string) *Machine {
	bs := []byte(s)
	ss := make(states, 0, len(bs)+1)
	for i, b := range bs {
		ss = append(ss, stateTo(b, i+1))
	}
	ss = append(ss, finalState())
	return &Machine{ss}
}

func Between(lo, hi rune) *Machine {
	if lo > hi {
		lo, hi = hi, lo
	}
	if lo < 0 || hi > 0x10ffff {
		panic("invalid range for unicode point")
	}
	u := &u8s{}
	u.between(lo, hi)
	return u.m()
}

func BetweenByte(s, e byte) *Machine {
	return &Machine{states{
		stateBetween(s, e, 1),
		finalState(),
	}}
}

func Char(s string) (m *Machine) {
	for _, r := range s {
		if r == utf8.RuneError {
			panic("invalid rune")
		}
		m = or2(m, Between(r, r))
	}
	return m
}

func opMany(op func(_, _ *Machine) *Machine, ms []*Machine) *Machine {
	switch len(ms) {
	case 0:
		return nil
	case 1:
		return ms[0].clone()
	}
	m := ms[0].clone()
	for i := 1; i < len(ms); i++ {
		m = op(m, ms[i])
	}
	return m
}

func Con(ms ...*Machine) *Machine {
	return opMany(con2, ms)
}
func con2(m1, m2 *Machine) *Machine {
	m := m1.clone()
	m2 = m2.clone()
	m2.shiftID(m.states.count() - 1)
	m.eachFinal(func(f *state) {
		f.connect(m2.startState())
		if !m2.startState().final() {
			f.label = notFinal
		}
	})
	m.states = append(m.states, m2.states[1:]...)
	return m
}

func Or(ms ...*Machine) *Machine {
	return opMany(or2, ms)
}
func or2(m1, m2 *Machine) *Machine {
	return newMerger(m1, m2, union{}).merge()
}

func And(ms ...*Machine) *Machine {
	return opMany(and2, ms)
}
func and2(m1, m2 *Machine) *Machine {
	return newMerger(m1, m2, intersection{}).merge()
}

func (m *Machine) ZeroOrMore() *Machine {
	m = m.OneOrMore()
	if len(m.states) == 2 {
		m.states = m.states[1:]
		m.shiftID(-1)
	}
	m.startState().label = defaultFinal
	return m.minimize()
}

func ZeroOrMore(ms ...*Machine) *Machine {
	return Con(ms...).ZeroOrMore()
}

func (m *Machine) OneOrMore() *Machine {
	m = m.clone()
	m.eachFinal(func(f *state) {
		f.connect(m.startState())
	})
	return m.minimize()
}

func OneOrMore(ms ...*Machine) *Machine {
	return Con(ms...).OneOrMore()
}

func (m *Machine) ZeroOrOne() *Machine {
	m = m.clone()
	m.states[0].label = defaultFinal
	return m.minimize()
}

func ZeroOrOne(ms ...*Machine) *Machine {
	return Con(ms...).ZeroOrOne()
}

func (m *Machine) Complement() *Machine {
	m = m.clone()
	m.each(func(f *state) {
		if f.final() {
			f.label = notFinal
		} else {
			f.label = defaultFinal
		}
	})
	return m.minimize()
}

func (m *Machine) Exclude(ms ...*Machine) *Machine {
	return and(m, Or(ms...).Complement())
}
