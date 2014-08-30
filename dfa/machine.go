package dfa

type Machine struct {
	states
}
type states []state

func (m *Machine) clone() *Machine {
	return &Machine{m.states.clone()}
}

func (m *Machine) startState() *state {
	return &m.states[0]
}

func (m *Machine) As(label int) *Machine {
	m.states.eachFinal(func(f *state) {
		f.label = finalLabel(label).toInternal()
	})
	return m
}

func (m *Machine) Match(src []byte) (size, label int, matched bool) {
	var (
		s   *state
		sid = stateID(0)
		p   = 0
	)
	for sid.valid() {
		s = &m.states[sid]
		if p < len(src) {
			sid = s.next(src[p])
			if sid.valid() {
				p++
			}
		} else {
			break
		}
	}
	if s.final() {
		return p, s.label.toExternal(), true
	}
	return p, -1, false
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
