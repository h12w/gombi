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

// Match greedily matches the DFA against src.
func (m *Machine) Match(src []byte) (size, label int, matched bool) {
	var (
		s, matchedState *state
		sid             = 0
		pos, matchedPos int
	)
	for sid >= 0 {
		s = &m.states[sid]
		if s.final() {
			matchedState = s
			matchedPos = pos
		}
		if pos < len(src) {
			sid = s.next(src[pos])
			if sid >= 0 {
				pos++
			}
		} else {
			break
		}
	}
	if matchedState != nil {
		return matchedPos, matchedState.label.toExternal(), true
	}
	return 0, -1, false
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

func (ss states) state(id int) *state {
	if id == -1 {
		return nil
	}
	return &ss[id]
}

func (ss states) shiftID(offset int) {
	ss.each(func(s *state) {
		s.each(func(t *trans) {
			t.next += offset
		})
	})
}
