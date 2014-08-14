package parse

type Token struct {
	Value interface{}
	*R
}

type Parser struct {
	cur, next *stateSet
}

func NewParser(r *R) *Parser {
	return &Parser{
		newStateSet(r.appendEOF()),
		newStateSet(nil)}
}
func (r *R) appendEOF() *R {
	for _, a := range r.Alts {
		if a.last() != EOF {
			a.Rules = append(a.Rules, EOF)
		}
	}
	return r
}

func (p *Parser) Parse(token *Token) {
	p.cur.each(func(s *state) {
		p.scanPredict(s, newTermState(token))
	})
	//fmt.Printf("set -> %s\n", p.cur.String())
	p.shift()
}
func (c *Parser) shift() {
	c.cur, c.next = c.next, newStateSet(nil)
}

func (ctx *Parser) scanPredict(s, t *state) {
	if !ctx.next.scan(s, t) {
		s.nextChildRule().eachAlt(func(alt *Alt) {
			if alt.isNull() {
				// copied because other alternatives should not be skipped
				ctx.scanPredict(s.copy().step(), t)
			} else if child := newState(alt); ctx.cur.add(child, s) {
				ctx.scanPredict(child, t)
			}
		})
	}
}

func (ss *stateSet) scan(s, t *state) bool {
	if s.scan(t) {
		if s.complete() && !s.last().isEOF() {
			for _, parent := range s.parents {
				// copied because multiple alternatives shares the same parent
				ss.scan(parent.copy(), s)
			}
		} else {
			ss.add(s, nil)
		}
		return true
	}
	return false
}

func (p *Parser) Results() (rs []*Node) {
	p.cur.each(func(s *state) {
		if s.complete() && s.last().isEOF() {
			rs = append(rs, newNode(s))
		}
	})
	return
}
