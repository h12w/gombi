package parse

type context struct {
	cur, next *stateSet
}

func newContext(r *rule) *context {
	cur, next := &stateSet{}, &stateSet{}
	cur.add(newState(r, 0))
	return &context{cur, next}
}

func (c *context) shift() {
	c.cur, c.next = c.next, &stateSet{}
}

func (ctx *context) scanPredict(token *token, s *state) {
	if !s.complete() {
		if ns, ok := s.scan(token); ok { // scan
			if !ns.complete() || ns.last().isEOF() {
				ctx.next.add(ns)
			}
			ctx.next.complete(ns) // complete
		} else { // predict
			ctx.cur.predict(token, s)
		}
	}
}

func (ss *stateSet) complete(s *state) {
	if s.complete() {
		s.parents.each(func(parent *state) {
			ns := parent.copy()
			ns.addChild(s)
			ns.d++
			if !ns.complete() || ns.last().isEOF() {
				ss.add(ns)
			}
			ss.complete(ns)
		})
	}
}

func (ss *stateSet) predict(token *token, s *state) {
	next := s.next()
	for j := range next.alts {
		ss.add(newState(next, j)).addParent(s)
	}
}
