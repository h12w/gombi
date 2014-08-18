package parse

type Parser struct {
	r         *R
	cur, next *stateSet
	results   []*Node
}

func NewParser(r *R) *Parser {
	r.appendEOF()
	p := &Parser{r: r}
	p.Reset()
	return p
}
func (p *Parser) Reset() {
	p.results = nil
	p.cur = newStateSet(p.r)
	p.next = newStateSet(nil)
	p.cur.each(func(s *state) {
		p.cur.predict(s)
	})
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
		p.scan(p.next, s, newTermState(t, r))
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

func (p *Parser) scan(ss *stateSet, s, t *state) bool {
	if s.scan(t) {
		if s.complete() {
			if s.last().isEOF() {
				p.results = append(p.results, newNode(s))
			} else {
				for _, parent := range s.parents {
					// copied because multiple alternatives shares the same parent
					p.scan(ss, parent.copy(), s)
				}
			}
		} else {
			ss.predict(s)
		}
		return true
	}
	return false
}

func (ss *stateSet) predict(s *state) {
	ss.add(s, nil) // NOTE from predictNull, s may not be added
	s.nextChildRule().eachAlt(func(alt *Alt) {
		if alt.isNull() {
			// copied because other alternatives should not be skipped
			ss.predictNull(s.copy().step())
		} else if child := newState(alt); ss.add(child, s) {
			ss.predict(child)
		}
	})
}

func (ss *stateSet) predictNull(s *state) {
	if s.complete() {
		for _, parent := range s.parents {
			np := parent.copy()
			np.scan(s)
			ss.predictNull(np)
		}
	} else {
		ss.predict(s)
	}
}

func (p *Parser) Results() (rs []*Node) {
	return p.results
}
