package dfa

import "container/list"

type merger struct {
	m1, m2, m *machine
	l         *list.List
	idm       map[[2]stateID]stateID
}

func newMerger(m1, m2 *machine) *merger {
	m := &merger{
		m1:  m1,
		m2:  m2,
		m:   &machine{},
		l:   list.New(),
		idm: make(map[[2]stateID]stateID)}
	m.getID(0, 0)
	return m
}

func (q *merger) merge() *machine {
	for q.l.Len() > 0 {
		id, id1, id2 := q.get()
		q.m.ss[id] = q.mergeState(q.m1.state(id1), q.m2.state(id2))
	}
	return q.m
}

func (q *merger) mergeState(s1, s2 *state) state {
	a := newTransArray()
	unionEachEdge(s1, s2, func(b byte, id1, id2 stateID) {
		a[b] = q.getID(id1, id2)
	})
	return state{a.toTransTable(), unionFinalLabel(s1, s2)}
}
func unionFinalLabel(s1, s2 *state) finalLabel {
	if s1 == nil {
		return s2.label
	}
	if s2 == nil {
		return s1.label
	}
	f1, f2 := s1.label, s2.label
	if f1 > defaultFinal && f2 > defaultFinal && f1 != f2 {
		panic("confilict label")
	}
	return finalMax(f1, f2)
}

func unionEachEdge(s1, s2 *state, visit func(b byte, id1, id2 stateID)) {
	it1, it2 := s1.iter(), s2.iter()
	b1, next1 := it1()
	b2, next2 := it2()
	for {
		b := b1
		id1, id2 := next1, next2
		if id1 == invalidID && id2 == invalidID {
			break
		} else if id1 == invalidID {
			b = b2
			b2, next2 = it2()
		} else if id2 == invalidID {
			b = b1
			b1, next1 = it1()
		} else {
			if b1 == b2 {
				b1, next1 = it1()
				b2, next2 = it2()
			} else if b1 < b2 {
				id2 = invalidID
				b1, next1 = it1()
			} else {
				b = b2
				id1 = invalidID
				b2, next2 = it2()
			}
		}
		visit(b, id1, id2)
	}
}

func (q *merger) getKey(id1, id2 stateID) [2]stateID {
	const trivialFinalID = -2
	if id1.valid() && q.m1.ss[id1].trivialFinal() {
		id1 = trivialFinalID
		if id2 == invalidID {
			id2 = trivialFinalID
		}
	}
	if id2.valid() && q.m2.ss[id2].trivialFinal() {
		id2 = trivialFinalID
		if id1 == invalidID {
			id1 = trivialFinalID
		}
	}
	return [2]stateID{id1, id2}
}
func (s *state) trivialFinal() bool {
	return s.label == defaultFinal && len(s.tt) == 0
}

func (q *merger) getID(id1, id2 stateID) stateID {
	key := q.getKey(id1, id2)
	if id, ok := q.idm[key]; ok {
		return id
	}
	id := stateID(len(q.m.ss))
	q.idm[key] = id
	q.m.ss = append(q.m.ss, state{})
	q.put(id, id1, id2)
	return id
}

func (q *merger) put(id, id1, id2 stateID) {
	q.l.PushFront([3]stateID{id, id1, id2})
}

func (q *merger) get() (id, id1, id2 stateID) {
	v := q.l.Remove(q.l.Back()).([3]stateID)
	return v[0], v[1], v[2]
}

func finalMax(a, b finalLabel) finalLabel {
	if a > b {
		return a
	}
	return b
}
