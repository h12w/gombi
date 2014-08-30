package dfa

func Str(s string) *machine {
	bs := []byte(s)
	ss := make(states, 0, len(bs)+1)
	for i, b := range bs {
		ss = append(ss, stateTo(b, stateID(i+1)))
	}
	ss = append(ss, finalState())
	return &machine{ss}
}

func Between(s, e byte) *machine {
	return &machine{states{
		stateBetween(s, e, 1),
		finalState(),
	}}
}

// TODO: unicode
func Char(s string) *machine {
	a := newTransArray()
	for i := range s {
		a.set(s[i], 1)
	}
	return &machine{states{
		a.toState(),
		finalState(),
	}}
}

func Con(ms ...*machine) *machine {
	if len(ms) == 0 {
		panic("zero machines")
	}
	m := ms[0]
	for i := 1; i < len(ms); i++ {
		m = con2(m, ms[i])
	}
	return m
}
func con2(m1, m2 *machine) *machine {
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

func Or(ms ...*machine) *machine {
	if len(ms) == 0 {
		panic("zero machines")
	}
	m := ms[0]
	for i := 1; i < len(ms); i++ {
		m = or2(m, ms[i])
	}
	return m
}
func or2(m1, m2 *machine) *machine {
	return newMerger(m1, m2).merge()
}

func ZeroOrMore(m *machine) *machine {
	m = OneOrMore(m)
	if len(m.states) == 2 {
		m.states = m.states[1:]
		m.shiftID(-1)
	}
	m.startState().label = defaultFinal
	return m
}

func OneOrMore(m *machine) *machine {
	m = m.clone()
	m.eachFinal(func(f *state) {
		f.connect(m.startState())
	})
	return m
}
