package parse

type Parser struct {
	r         *R
	cur, next states
	pset      stateSet
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
	p.cur.reset()
	p.next.reset()
	p.pset.reset()
	p.r.eachAlt(func(a *Alt) {
		p.predict(newState(a))
	})
	p.shift()
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
	//fmt.Println("scanning", newTermState(t, r))
	//fmt.Println()
	p.cur.each(func(s *state) {
		p.scan(s, newTermState(t, r))
	})
	p.shift()
	return true
}
func (p *Parser) shift() {
	//fmt.Printf("predict set -> %s\n", p.pset.String())
	//fmt.Println()
	//fmt.Printf("terminal set -> %s\n", p.next.String())
	//fmt.Println()
	p.cur, p.next = p.next, p.cur.reset()
	p.pset.reset()
}

func (p *Parser) scan(s, t *state) bool {
	if s.scan(t) {
		if s.complete() {
			if s.last().isEOF() {
				p.results = append(p.results, newNode(s))
			} else {
				for _, parent := range s.parents {
					// copied because multiple alternatives shares the same parent
					p.scan(parent.copy(), s)
				}
			}
		} else {
			p.predict(s) // NOTE predict in the next set
		}
		return true
	}
	return false
}

func (p *Parser) predict(s *state) {
	if s.isTerm() {
		p.next.append(s)
		return
	}
	s.nextChildRule().eachAlt(func(alt *Alt) { // NOTE term rule does not have alt
		if alt.isNull() {
			// copied because other alternatives should not be skipped
			p.predictNull(s.copy().step())
		} else if child := newState(alt); p.pset.add(child, s) {
			p.predict(child)
		}
	})
}

func (p *Parser) predictNull(s *state) {
	if s.complete() {
		for _, parent := range s.parents {
			np := parent.copy()
			np.scan(s)
			p.predictNull(np)
		}
	} else {
		p.predict(s)
	}
}

func (p *Parser) Results() (rs []*Node) {
	return p.results
}
