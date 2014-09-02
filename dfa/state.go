package dfa

type state struct {
	table transTable
	label finalLabel
}
type transTable []trans
type trans struct {
	s, e byte
	next int
}
type transArray [256]int

func stateBetween(lo, hi byte, next int) state {
	a := newTransArray()
	a.setBetween(lo, hi, next)
	return state{table: a.toTransTable()}
}

func stateTo(b byte, next int) state {
	a := newTransArray()
	a.set(b, next)
	return state{table: a.toTransTable()}
}

func finalState() state {
	return state{label: defaultFinal}
}

func (s *state) final() bool {
	return s.label.final()
}

func (s *state) connect(o *state) {
	a := s.table.toTransArray()
	o.each(func(t *trans) {
		a.setBetween(t.s, t.e, t.next)
	})
	s.table = a.toTransTable()
}

func (s *state) clone() state {
	return state{s.table.clone(), s.label}
}

func (s *state) iter() func() (byte, int) {
	if s == nil || len(s.table) == 0 {
		return func() (byte, int) {
			return 0, invalidID
		}
	}
	i := 0
	b := s.table[i].s
	return func() (byte, int) {
		defer func() {
			if i < len(s.table) {
				if b == s.table[i].e {
					i++
					if i < len(s.table) {
						b = s.table[i].s
					}
				} else {
					b++
				}
			}
		}()
		if i < len(s.table) {
			return b, s.table[i].next
		}
		return 0, invalidID
	}
}

func (s *state) each(visit func(*trans)) {
	s.table.each(visit)
}

func (s *state) next(b byte) (sid int) {
	for i := range s.table {
		if s.table[i].s <= b && b <= s.table[i].e {
			return s.table[i].next
		}
	}
	return invalidID
}

func (s *state) binaryNext(b byte) (sid int) {
	min, max := 0, len(s.table)-1
	for min <= max {
		mid := (min + max) / 2
		if s.table[mid].s <= b && b <= s.table[mid].e {
			return s.table[mid].next
		} else if b < s.table[mid].s {
			max = mid - 1
		} else {
			min = mid + 1
		}
	}
	return invalidID
}

func (table *transTable) each(visit func(*trans)) {
	for i := range *table {
		visit(&(*table)[i])
	}
}

func (table *transTable) clone() transTable {
	return append(transTable(nil), *table...)
}

func (table *transTable) toTransArray() transArray {
	a := newTransArray()
	table.each(func(t *trans) {
		for i := t.s; i <= t.e; i++ {
			a[i] = t.next
		}
	})
	return a
}

func newTransArray() (a transArray) {
	for i := range a {
		a[i] = invalidID
	}
	return a
}

func (a *transArray) set(b byte, next int) *transArray {
	if a[b] == invalidID {
		a[b] = next
		return a
	}
	if a[b] != next {
		panic("a different trans already set")
	}
	return a
}

func (a *transArray) setBetween(lo, hi byte, next int) *transArray {
	b := lo
	for {
		a.set(b, next)
		if b == hi {
			break
		} else {
			b++
		}
	}
	return a
}

func (a *transArray) toTransTable() (table transTable) {
	i := 0
	for ; i < len(a); i++ {
		if a[i] != invalidID {
			break
		}
	}
	if i == 256 {
		return
	}
	table = append(table, trans{byte(i), byte(i), a[i]})
	i++
	for ; i < len(a); i++ {
		if a[i] != invalidID {
			b := byte(i)
			last := table[len(table)-1]
			if b == last.e+1 && a[i] == last.next {
				table[len(table)-1].e = b
			} else {
				table = append(table, trans{b, b, a[i]})
			}
		}
	}
	return
}

func (a *transArray) toState() state {
	return state{table: a.toTransTable()}
}
