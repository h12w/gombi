package parse

type Token struct {
	Value interface{}
	*R
}

type Parser struct {
	cur, next *stateSet
}

func NewParser(r *R) *Parser {
	cur, next := newStateSet(), newStateSet()
	// TODO: detect and add EOF
	cur.add(newState(r, r.Alts[0])) // TODO: Alts[0] needs improvement
	return &Parser{cur, next}
}

func (p *Parser) Parse(token *Token) {
	p.cur.each(func(s *state) {
		p.scanPredict(s, newTermState(token))
	})
	//fmt.Printf("set -> %s\n", p.cur.String())
	p.shift()
}

func (p *Parser) Result() *Node {
	if len(p.cur.a) > 0 {
		return &Node{p.cur.a[0]}
	}
	return nil
}

func (c *Parser) shift() {
	c.cur, c.next = c.next, newStateSet()
}

func (ctx *Parser) scanPredict(s, t *state) {
	if !ctx.next.scan(s, t) {
		s.nextChildRule().eachAlt(func(r *R, alt *Alt) {
			if alt.isNull() {
				ns := s.copy()
				ns.d++
				ctx.scanPredict(ns, t)
			} else {
				child, isNew := ctx.cur.add(newState(r, alt))
				child.parents.add(s)
				if isNew {
					ctx.scanPredict(child, t)
				}
			}
		})
	}
}

func (ss *stateSet) scanNull(s, t *state) bool {
	if s.R == Null {
		ss.scan(s, newTermState(&Token{R: Null}))
	}
	return ss.scan(s, t)
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
