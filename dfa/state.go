package dfa

// final labels
const (
	notFinal     = 0
	defaultFinal = 1
)

type state struct {
	tt         transTable
	finalLabel int
}
type transTable []trans
type trans struct {
	s, e byte
	next int
}
type transArray [256]int

func newTransArray() (a transArray) {
	for i := range a {
		a[i] = -1
	}
	return
}

func (s *state) final() bool {
	return s.finalLabel > notFinal
}

func (s *state) trivialFinal() bool {
	return s.finalLabel == defaultFinal && len(s.tt) == 0
}

func (s *state) clone() state {
	return state{s.tt.clone(), s.finalLabel}
}

func (s *state) set(b byte, next int) {
	if a := s.tt.toTransArray(); a[b] == -1 {
		a[b] = next
		s.tt = a.toTransTable()
		return
	}
	panic("trans already set")
}

func (s *state) setBetween(from, to byte, next int) {
	a := s.tt.toTransArray()
	for b := from; b <= to; b++ {
		if a[b] == -1 {
			a[b] = next
		} else {
			panic("trans already set")
		}
	}
	s.tt = a.toTransTable()
}

func (s *state) iter() func() (byte, int) {
	if s == nil || len(s.tt) == 0 {
		return func() (byte, int) {
			return 0, -1
		}
	}
	i := 0
	b := s.tt[i].s
	return func() (byte, int) {
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
		return 0, -1
	}
}

func (s *state) each(visit func(*trans)) {
	s.tt.each(visit)
}

func (s *state) next(b byte) (sid int) {
	for i := range s.tt {
		if s.tt[i].s <= b && b <= s.tt[i].e {
			return s.tt[i].next
		}
	}
	return -1
}

func (s *state) binaryNext(b byte) (sid int) {
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
	return -1
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

func (ts transArray) toTransTable() (tt transTable) {
	i := 0
	for ; i < len(ts); i++ {
		if ts[i] != -1 {
			break
		}
	}
	if i == 256 {
		return
	}
	tt = append(tt, trans{byte(i), byte(i), ts[i]})
	i++
	for ; i < len(ts); i++ {
		if ts[i] != -1 {
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
