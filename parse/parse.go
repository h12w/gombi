package parse

type Parser struct {
	r         *R
	cur, next states
	results   []*Node
	lastTok   *R
}

func New(r *R) *Parser {
	r.appendEOF()
	p := &Parser{r: r}
	p.Reset()
	return p
}
func (p *Parser) Reset() {
	p.results = nil
	p.cur.reset()
	p.next.reset()
	//pset := &stateSet{}
	//p.cur.each(func(s *state) {
	//	p.next.predictAll(pset, s, nil) // TODO: not unique
	//})
	//p.Parse(nil, SOF)
}
func (r *R) appendEOF() *R {
	if r.recursive {
		return con(r, EOF)
		//return con(SOF, r, EOF)
	}
	for _, a := range r.Alts {
		if a.last() != EOF {
			//a.Rules = append(Rules{SOF}, a.Rules...)
			a.Rules = append(a.Rules, EOF)
		}
	}
	return r
}

func (p *Parser) Parse(t *Token, tr *R) (ok bool) {
	//fmt.Println("scanning", newTermState(t, r))
	//fmt.Println()
	pset := &stateSet{}
	if p.lastTok == nil {
		p.r.eachAlt(func(alt *Alt) {
			p.cur.predictAll(pset, newState(alt), tr)
		})
	}
	if len(p.cur.a) == 0 {
		return false
	}
	p.cur.each(func(s *state) {
		if s.rule() == p.lastTok || p.lastTok == nil {
			p.next.predictAll(pset, s, tr)
		}
	})
	p.shift()
	pset = &stateSet{}
	p.cur.each(func(s *state) {
		if s.scan(t, tr) {
			ok = true
			p.results = append(p.results, s.propagate()...)
		}
	})
	p.lastTok = tr
	//fmt.Println("### token ->", tr)
	//fmt.Printf("### cur set ->\n%s\n", p.cur.dumpUp())
	//fmt.Println()
	//fmt.Printf("### predict set ->\n%s\n", pset.String())
	//fmt.Println()
	return
}
func (p *Parser) shift() {
	p.cur, p.next = p.next, p.cur.reset()
}

func (p *Parser) Error() error {
	return nil
}

func (p *Parser) Results() (rs []*Node) {
	return p.results
}

func (ss *states) predictAll(pset *stateSet, s *state, termRule *R) {
	if s.complete() {
		for _, parent := range s.parents {
			ss.predictAll(pset, parent.advance(s), termRule)
		}
	} else {
		ss.predict(pset, s, termRule)
	}
	return
}

func (s *state) propagate() (results []*Node) {
	if s.complete() {
		if s.last().isEOF() {
			return []*Node{newNode(s)}
		} else {
			for _, parent := range s.parents {
				results = append(results, parent.advance(s).propagate()...)
			}
		}
	}
	return
}

func (ss *states) predict(pset *stateSet, s *state, termRule *R) {
	if s.isTerm {
		if termRule == nil || termRule == s.rule() {
			ss.append(s)
		}
		return
	}
	s.nextChildRule().eachAlt(func(alt *Alt) {
		if alt.isNull() {
			// copied because other alternatives should not be skipped
			ss.predictNull(pset, s.copy().step(), termRule)
		} else if child, isNew := pset.add(alt, s); isNew {
			ss.predict(pset, child, termRule)
		}
	})
}

func (ss *states) predictNull(pset *stateSet, s *state, termRule *R) {
	if s.complete() {
		for _, parent := range s.parents {
			ss.predictNull(pset, parent.advance(s), termRule)
		}
	} else {
		ss.predict(pset, s, termRule)
	}
}
