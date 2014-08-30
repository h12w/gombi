package dfa

func Str(s string) *machine {
	bs := []byte(s)
	ss := make([]state, len(bs)+1)
	for i := range bs {
		ss[i].set(bs[i], stateID(i+1))
	}
	ss[len(ss)-1].label = defaultFinal
	return &machine{ss}
}

func Between(s, e byte) *machine {
	ss := make([]state, 2)
	ss[0].setBetween(s, e, 1)
	ss[1].label = defaultFinal
	return &machine{ss}
}

// TODO: unicode
func Char(s string) *machine {
	a := newTransArray()
	for i := range s {
		a.set(s[i], 1)
	}
	ss := make([]state, 2)
	ss[0].tt = a.toTransTable()
	ss[1].label = defaultFinal
	return &machine{ss}
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
	m2 = m2.clone().shiftID(m.stateCount() - 1)
	m.each(func(s *state) {
		if s.final() {
			s.connect(&m2.ss[0])
			if !m2.ss[0].final() {
				s.label = notFinal
			}
		}
	})
	m.ss = append(m.ss, m2.ss[1:]...)
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
	if len(m.ss) == 2 {
		m.ss = m.ss[1:]
		m.shiftID(-1)
	}
	m.ss[0].label = defaultFinal
	return m
}

func OneOrMore(m *machine) *machine {
	m = m.clone()
	m.each(func(s *state) {
		if s.final() {
			s.connect(&m.ss[0])
		}
	})
	return m
}
