package dfa

func Str(s string) *Machine {
	bs := []byte(s)
	ss := make(states, 0, len(bs)+1)
	for i, b := range bs {
		ss = append(ss, stateTo(b, i+1))
	}
	ss = append(ss, finalState())
	return &Machine{ss}
}

func Between(s, e byte) *Machine {
	return &Machine{states{
		stateBetween(s, e, 1),
		finalState(),
	}}
}

// TODO: unicode
func Char(s string) *Machine {
	a := newTransArray()
	for i := range s {
		a.set(s[i], 1)
	}
	return &Machine{states{
		a.toState(),
		finalState(),
	}}
}

func Con(ms ...*Machine) *Machine {
	if len(ms) == 0 {
		panic("zero Machines")
	}
	m := ms[0]
	for i := 1; i < len(ms); i++ {
		m = con2(m, ms[i])
	}
	return m
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
	if len(ms) == 0 {
		panic("zero Machines")
	}
	m := ms[0]
	for i := 1; i < len(ms); i++ {
		m = or2(m, ms[i])
	}
	return m
}
func or2(m1, m2 *Machine) *Machine {
	return newMerger(m1, m2).merge()
}

func ZeroOrMore(m *Machine) *Machine {
	m = OneOrMore(m)
	if len(m.states) == 2 {
		m.states = m.states[1:]
		m.shiftID(-1)
	}
	m.startState().label = defaultFinal
	return m
}

func OneOrMore(m *Machine) *Machine {
	m = m.clone()
	m.eachFinal(func(f *state) {
		f.connect(m.startState())
	})
	return m
}
