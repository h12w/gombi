package dfa

type machine struct {
	ss []state
}

func (m *machine) clone() *machine {
	ss := make([]state, len(m.ss))
	for i := range ss {
		ss[i] = m.ss[i].clone()
	}
	return &machine{ss}
}

func (m *machine) each(visit func(*state)) {
	for i := range m.ss {
		visit(&m.ss[i])
	}
}

func (m *machine) shiftID(offset int) *machine {
	m.each(func(s *state) {
		s.each(func(t *trans) {
			t.next += offset
		})
	})
	return m
}

func (m *machine) state(id int) *state {
	if id == -1 {
		return nil
	}
	return &m.ss[id]
}

func (m *machine) stateCount() int {
	return len(m.ss)
}

func (m *machine) startState() *state {
	return &m.ss[0]
}
