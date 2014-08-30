package dfa

type machine struct {
	states
}
type states []state

func (m *machine) clone() *machine {
	return &machine{m.states.clone()}
}

func (m *machine) startState() *state {
	return &m.states[0]
}

func (m *machine) withLabel(label finalLabel) *machine {
	m.states.eachFinal(func(f *state) {
		f.label = label
	})
	return m
}

func (ss states) each(visit func(*state)) {
	for i := range ss {
		visit(&ss[i])
	}
}

func (ss states) eachFinal(visit func(*state)) {
	for i := range ss {
		if ss[i].final() {
			visit(&ss[i])
		}
	}
}

func (ss states) count() int {
	return len(ss)
}

func (ss states) clone() states {
	ss = append(states(nil), ss...)
	for i := range ss {
		ss[i] = ss[i].clone()
	}
	return ss
}

func (ss states) state(id stateID) *state {
	if id == -1 {
		return nil
	}
	return &ss[id]
}

func (ss states) shiftID(offset int) {
	ss.each(func(s *state) {
		s.each(func(t *trans) {
			t.next += stateID(offset)
		})
	})
}
