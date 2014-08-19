package parse

type Parser struct {
	r         *R
	cur, next states
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
	p.r.eachAlt(func(a *Alt) {
		p.cur.predict(&stateSet{}, newState(a))
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

func (p *Parser) Parse(t *Token, tr *R) bool {
	//fmt.Println("scanning", newTermState(t, r))
	//fmt.Println()
	pset := &stateSet{}
	p.cur.each(func(s *state) {
		if s.scan(t, tr) {
			p.results = append(p.results, p.next.propagate(pset, s)...)
		}
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
}

func (p *Parser) Error() error {
	return nil
}

func (p *Parser) Results() (rs []*Node) {
	return p.results
}

func (ss *states) propagate(pset *stateSet, s *state) (results []*Node) {
	if s.complete() {
		if s.last().isEOF() {
			return []*Node{newNode(s)}
		} else {
			for _, parent := range s.parents {
				results = append(results, ss.propagate(pset, parent.advance(s))...)
			}
		}
	} else {
		ss.predict(pset, s)
	}
	return
}

func (ss *states) predict(pset *stateSet, s *state) {
	if s.isTerm() {
		ss.append(s)
		return
	}
	s.nextChildRule().eachAlt(func(alt *Alt) {
		if alt.isNull() {
			// copied because other alternatives should not be skipped
			ss.predictNull(pset, s.copy().step())
		} else if child := newState(alt); pset.add(child, s) {
			ss.predict(pset, child)
		}
	})
}

func (ss *states) predictNull(pset *stateSet, s *state) {
	if s.complete() {
		for _, parent := range s.parents {
			ss.predictNull(pset, parent.advance(s))
		}
	} else {
		ss.predict(pset, s)
	}
}
