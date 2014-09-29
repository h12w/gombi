package parse

type Parser struct {
	r         *R
	cur, next states
	results   []*Node
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
	m := make(map[*R]bool)
	p.r.eachAlt(func(a *Alt) {
		if !m[a.R] {
			p.cur.predict(&stateSet{}, newState(a)) // TODO: not unique
			m[a.R] = true
		}
	})
}
func (r *R) appendEOF() *R {
	if r.recursive {
		return con(r, EOF)
	}
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
	//fmt.Println("### token ->", tr)
	//fmt.Printf("### cur set ->\n%s\n", p.cur.dumpUp())
	//fmt.Println()
	//fmt.Printf("### predict set ->\n%s\n", pset.String())
	//fmt.Println()
	p.shift()
	return true
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
	if s.isTerm {
		ss.append(s)
		return
	}
	s.nextChildRule().eachAlt(func(alt *Alt) {
		if alt.isNull() {
			// copied because other alternatives should not be skipped
			ss.predictNull(pset, s.copy().step())
		} else if child, isNew := pset.add(alt, s); isNew {
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
