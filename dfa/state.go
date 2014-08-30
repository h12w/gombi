package dfa

type finalLabel int

const (
	notFinal     finalLabel = 0
	defaultFinal finalLabel = 1
)

type stateID int

const (
	invalidID      stateID = -1
	trivialFinalID stateID = -2
)

func (id stateID) valid() bool {
	return id >= 0
}

type state struct {
	tt    transTable
	label finalLabel
}
type transTable []trans
type trans struct {
	s, e byte
	next stateID
}
type transArray []stateID

func (s *state) final() bool {
	return s.label > notFinal
}

func (s *state) set(b byte, next stateID) {
	s.tt = s.tt.toTransArray().set(b, next).toTransTable()
}

func (s *state) setBetween(from, to byte, next stateID) {
	s.tt = s.tt.toTransArray().setBetween(from, to, next).toTransTable()
}

func (s *state) connect(o *state) {
	a := s.tt.toTransArray()
	o.each(func(t *trans) {
		a.setBetween(t.s, t.e, t.next)
	})
	s.tt = a.toTransTable()
}

func (s *state) clone() state {
	return state{s.tt.clone(), s.label}
}

func (s *state) iter() func() (byte, stateID) {
	if s == nil || len(s.tt) == 0 {
		return func() (byte, stateID) {
			return 0, invalidID
		}
	}
	i := 0
	b := s.tt[i].s
	return func() (byte, stateID) {
		defer func() {
			if i < len(s.tt) {
				if b == s.tt[i].e {
					i++
					if i < len(s.tt) {
						b = s.tt[i].s
					}
				} else {
					b++
				}
			}
		}()
		if i < len(s.tt) {
			return b, s.tt[i].next
		}
		return 0, invalidID
	}
}

func (s *state) each(visit func(*trans)) {
	s.tt.each(visit)
}

func (s *state) next(b byte) (sid stateID) {
	for i := range s.tt {
		if s.tt[i].s <= b && b <= s.tt[i].e {
			return s.tt[i].next
		}
	}
	return invalidID
}

func (s *state) binaryNext(b byte) (sid stateID) {
	min, max := 0, len(s.tt)-1
	for min <= max {
		mid := (min + max) / 2
		if s.tt[mid].s <= b && b <= s.tt[mid].e {
			return s.tt[mid].next
		} else if b < s.tt[mid].s {
			max = mid - 1
		} else {
			min = mid + 1
		}
	}
	return invalidID
}

func (tt *transTable) each(visit func(*trans)) {
	for i := range *tt {
		visit(&(*tt)[i])
	}
}

func (tt *transTable) clone() transTable {
	return append(transTable(nil), *tt...)
}

func (tt *transTable) toTransArray() transArray {
	a := newTransArray()
	tt.each(func(t *trans) {
		for i := t.s; i <= t.e; i++ {
			a[i] = t.next
		}
	})
	return a
}

func (t *trans) each(visit func(byte)) {
	b := t.s
	for {
		visit(b)
		if b == t.e {
			break
		}
		b++
	}
}

func newTransArray() transArray {
	a := make(transArray, 256)
	for i := range a {
		a[i] = invalidID
	}
	return a
}

func (a transArray) set(b byte, next stateID) transArray {
	if a[b] == invalidID {
		a[b] = next
		return a
	}
	panic("trans already set")
}

func (a transArray) setBetween(from, to byte, next stateID) transArray {
	for b := from; b <= to; b++ {
		a.set(b, next)
	}
	return a
}

func (ts transArray) toTransTable() (tt transTable) {
	i := 0
	for ; i < len(ts); i++ {
		if ts[i] != invalidID {
			break
		}
	}
	if i == 256 {
		return
	}
	tt = append(tt, trans{byte(i), byte(i), ts[i]})
	i++
	for ; i < len(ts); i++ {
		if ts[i] != invalidID {
			b := byte(i)
			last := tt[len(tt)-1]
			if b == last.e+1 && ts[i] == last.next {
				tt[len(tt)-1].e = b
			} else {
				tt = append(tt, trans{b, b, ts[i]})
			}
		}
	}
	return
}
