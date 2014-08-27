package dfa

type state struct {
	tt     transTable
	accept bool
}
type transTable []trans
type trans struct {
	s, e byte
	next *state
}
type transArray [256]*state

func (s *state) set(b byte, next *state) {
	if trans := s.tt.toTransArray(); trans[b] == nil {
		trans[b] = next
		s.tt = trans.toTransTable()
		return
	}
	panic("trans already set")
}

func (s *state) setBetween(from, to byte, next *state) {
	trans := s.tt.toTransArray()
	for b := from; b <= to; b++ {
		if trans[b] == nil {
			trans[b] = next
		} else {
			panic("trans already set")
		}
	}
	s.tt = trans.toTransTable()
}

func (s *state) find(b byte) *state {
	for i := range s.tt {
		if s.tt[i].s <= b && b <= s.tt[i].e {
			return s.tt[i].next
		}
	}
	return nil
}

func (s *state) next(b byte) *state {
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
	return nil
}

func (tt transTable) toTransArray() (a transArray) {
	for _, trans := range tt {
		for i := trans.s; i <= trans.e; i++ {
			a[i] = trans.next
		}
	}
	return
}

func (ts transArray) toTransTable() (tt transTable) {
	i := 0
	for ; i < len(ts); i++ {
		if ts[i] != nil {
			break
		}
	}
	tt = append(tt, trans{byte(i), byte(i), ts[i]})
	i++
	for ; i < len(ts); i++ {
		if ts[i] != nil {
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
