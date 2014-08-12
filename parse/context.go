package parse

type context struct {
	cur, next *stateSet
}

func newContext(r *R) *context {
	cur, next := newStateSet(), newStateSet()
	cur.add(newState(r, r.Alts[0])) // TODO: Alts[0] needs improvement
	return &context{cur, next}
}

func (c *context) shift() {
	c.cur, c.next = c.next, newStateSet()
}

func (ctx *context) scanPredict(s, t *state) {
	if !ctx.next.scan(s, t) {
		s.nextChildR().eachAlt(func(r *R, alt *Alt) {
			child, isNew := ctx.cur.add(newState(r, alt))
			child.parents.add(s)
			if isNew {
				ctx.scanPredict(child, t)
			}
		})
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
