package parse

type context struct {
	cur, next *stateSet
}

func newContext(r *rule) *context {
	cur, next := newStateSet(), newStateSet()
	cur.add(newState(r, 0))
	return &context{cur, next}
}

func (c *context) shift() {
	c.cur, c.next = c.next, newStateSet()
}

func (ctx *context) scanPredict(s, t *state) {
	if !ctx.next.scan(s, t) {
		next := s.next()
		for i := range next.alts {
			child, isNew := ctx.cur.add(newState(next, i))
			child.addParent(s)
			if isNew {
				ctx.scanPredict(child, t)
			}
		}
	}
}

func (ss *stateSet) scan(s, t *state) bool {
	if s.scan(t) { // scan
		if s.complete() && !s.last().isEOF() {
			s.parents.each(func(parent *state) {
				// multiple alternatives shares the same parent, so it must be copied
				ss.scan(parent.copy(), s)
			})
		} else {
			// add only not completed states or the final result
			ss.add(s)
		}
		return true
	}
	return false
}
