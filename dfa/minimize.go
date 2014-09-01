package dfa

func (m *Machine) minimize() *Machine {
	n := m.states.count()
	diff := newDiff(n)
	diff.eachSame(func(i, j int) {
		s, t := m.states[i], m.states[j]
		diff.set(i, j, s.label != t.label ||
			s.table.toTransSet() != t.table.toTransSet())
	})
	for diff.hasNewDiff {
		diff.hasNewDiff = false
		diff.eachSame(func(i, j int) {
			s, t := m.states[i], m.states[j]
			si, ti := s.iter(), t.iter()
			sb, sid := si()
			tb, tid := ti()
			if sb != tb {
				panic("assert: each iteration should return the same input byte")
			}
			for sid != -1 && tid != -1 {
				if sb != tb {
					panic("assert: each iteration should return the same input byte")
				}
				diff.set(i, j, sid != tid && diff.get(sid, tid))
				sb, sid = si()
				tb, tid = ti()
			}
		})
	}
	idm := make(map[int]int)
	diff.eachSame(func(i, j int) {
		idm[j] = i
	})
	m.each(func(s *state) {
		s.each(func(t *trans) {
			if small, ok := idm[t.next]; ok {
				t.next = small
			}
		})
	})
	return or2(m, m) // or2(m, m) is also a way to remove unreachable nodes
}

type transSet [256]bool

func (t *transTable) toTransSet() (s transSet) {
	t.each(func(t *trans) {
		for b := t.s; b <= t.e; b++ {
			s[b] = true
		}
	})
	return
}

// 0: 1, 2, ..., n-1
// 1:    2, ..., n-1
// ...
// n-2:          n-1
type diff struct {
	n          int
	a          []bool
	hasNewDiff bool
}

func newDiff(n int) *diff {
	return &diff{n, make([]bool, n*(n-1)/2), false}
}

func (d *diff) set(i, j int, different bool) {
	d.hasNewDiff = d.hasNewDiff || different
	if different {
		d.a[d.index(i, j)] = different
	}
}

func (d *diff) get(i, j int) bool {
	return d.a[d.index(i, j)]
}

func (d *diff) index(i, j int) int {
	if i == j {
		panic("i should not be equal to j")
	}
	if i > j {
		i, j = j, i
	}
	return (2*d.n-i-1)*i/2 + (j - i - 1)
}

func (d *diff) eachSame(visit func(int, int)) {
	for i := d.n - 2; i >= 0; i-- { // reverse order so the smaller comes later
		for j := i + 1; j <= d.n-1; j++ {
			if !d.get(i, j) {
				visit(i, j)
			}
		}
	}
}
