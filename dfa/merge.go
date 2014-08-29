package dfa

import "container/list"

type merger struct {
	m1, m2, m *machine
	l         *list.List
	idm       map[[2]int]int
}

func newMerger(m1, m2 *machine) *merger {
	m := &merger{
		m1:  m1,
		m2:  m2,
		m:   &machine{},
		l:   list.New(),
		idm: make(map[[2]int]int)}
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
	unionEachEdge(s1, s2, func(b byte, id1, id2 int) {
		a[b] = q.getID(id1, id2)
	})
	return state{a.toTransTable(), unionFinalID(s1, s2)}
}
func unionFinalID(s1, s2 *state) int {
	if s1 == nil {
		return s2.finalLabel
	}
	if s2 == nil {
		return s1.finalLabel
	}
	f1, f2 := s1.finalLabel, s2.finalLabel
	if f1 > defaultFinal && f2 > defaultFinal && f1 != f2 {
		panic("confilict finalLabel")
	}
	return iMax(f1, f2)
}

func unionEachEdge(s1, s2 *state, visit func(b byte, id1, id2 int)) {
	it1, it2 := s1.iter(), s2.iter()
	b1, next1 := it1()
	b2, next2 := it2()
	for {
		b := b1
		id1, id2 := next1, next2
		if id1 == -1 && id2 == -1 {
			break
		} else if id1 == -1 {
			b = b2
			b2, next2 = it2()
		} else if id2 == -1 {
			b = b1
			b1, next1 = it1()
		} else {
			if b1 == b2 {
				b1, next1 = it1()
				b2, next2 = it2()
			} else if b1 < b2 {
				id2 = -1
				b1, next1 = it1()
			} else {
				b = b2
				id1 = -1
				b2, next2 = it2()
			}
		}
		visit(b, id1, id2)
	}
}

func (q *merger) getKey(id1, id2 int) [2]int {
	const trivialFinalID = -2
	if id1 >= 0 && q.m1.ss[id1].trivialFinal() {
		id1 = trivialFinalID
		if id2 == -1 {
			id2 = trivialFinalID
		}
	}
	if id2 >= 0 && q.m2.ss[id2].trivialFinal() {
		id2 = trivialFinalID
		if id1 == -1 {
			id1 = trivialFinalID
		}
	}
	return [2]int{id1, id2}
}

func (q *merger) getID(id1, id2 int) int {
	key := q.getKey(id1, id2)
	if id, ok := q.idm[key]; ok {
		return id
	}
	id := len(q.m.ss)
	q.idm[key] = id
	q.m.ss = append(q.m.ss, state{})
	q.put(id, id1, id2)
	return id
}

func (q *merger) put(id, id1, id2 int) {
	q.l.PushFront([3]int{id, id1, id2})
}

func (q *merger) get() (id int, id1 int, id2 int) {
	v := q.l.Remove(q.l.Back()).([3]int)
	return v[0], v[1], v[2]
}

func iMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
