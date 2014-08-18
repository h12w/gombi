package parse

type Parser struct {
	r         *R
	cur, next *stateSet
}

func NewParser(r *R) *Parser {
	p := &Parser{r: r}
	p.Reset()
	return p
}
func (p *Parser) Reset() {
	p.cur = newStateSet(p.r.appendEOF())
	p.next = newStateSet(nil)
}
func (r *R) appendEOF() *R {
	for _, a := range r.Alts {
		if a.last() != EOF {
			a.Rules = append(a.Rules, EOF)
		}
	}
	return r
}

func (p *Parser) Error() error {
	return nil
}

func (p *Parser) Parse(t *Token, r *R) bool {
	//fmt.Printf("cur set -> %s\n", p.cur.String())
	p.cur.each(func(s *state) {
		p.scanPredict(s, newTermState(t, r))
	})
	//fmt.Printf("cur set -> %s\n", p.cur.String())
	//fmt.Printf("next set -> %s\n", p.next.String())
	//fmt.Println()
	p.shift()
	return true
}
func (c *Parser) shift() {
	c.cur, c.next = c.next, c.cur.reset()
}

func (ctx *Parser) scanPredict(s, t *state) {
	if !ctx.next.scan(s, t) {
		s.nextChildRule().eachAlt(func(alt *Alt) {
			if alt.isNull() {
				// copied because other alternatives should not be skipped
				ctx.scanNull(s.copy().step(), t)
			} else if child := newState(alt); ctx.cur.add(child, s) {
				ctx.scanPredict(child, t)
			}
		})
	}
}

func (ctx *Parser) scanNull(s, t *state) {
	if s.complete() {
		for _, parent := range s.parents {
			np := parent.copy()
			np.scan(s)
			ctx.scanNull(np, t)
		}
	} else {
		ctx.scanPredict(s, t)
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
			ss.add(s, nil) // TODO may not need to check duplicates
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
