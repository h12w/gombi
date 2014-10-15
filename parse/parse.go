package parse

type Parser struct {
	r       *R
	s       *state
	results []*Node
}

func New(r *R) *Parser {
	p := &Parser{r: r}
	p.Reset()
	return p
}
func (p *Parser) Reset() {
	p.results = nil
	p.s = nil
}

func (p *Parser) Parse(t *Token, tr *R) bool {
	pset := newStateSet(tr.Alts[0])
	if p.s == nil {
		p.r.eachAlt(func(alt *Alt) {
			pset.predictNext(newState(alt))
		})
	} else {
		pset.predict(p.s)
		p.s.parents = nil
	}
	p.s = pset.termState
	//fmt.Printf("### predict set ->\n%s\n", pset.String())
	//fmt.Println()
	//fmt.Printf("### term state ->\n%s\n", pset.termState.dumpUp(0))
	//fmt.Println()
	if p.s == nil {
		return false
	}
	p.s.scan(t)
	if tr == EOF {
		p.collectResult(p.s)
		return false
	}
	return true
}

func (p *Parser) Error() error {
	return nil
}

func (p *Parser) Results() []*Node {
	return p.results
}

func (p *Parser) collectResult(s *state) {
	if s.complete() {
		if s.rule() == p.r {
			p.results = append(p.results, newNode(s))
		}
		for _, parent := range s.parents {
			p.collectResult(parent.advance(s))
		}
		s.parents = nil // OK
	}
}

func (pset *stateSet) predict(s *state) {
	if s.complete() {
		for _, parent := range s.parents {
			pset.predict(parent.advance(s))
		}
		return
	}
	pset.predictNext(s)
}

func (pset *stateSet) predictNext(s *state) {
	if s.isTerm() {
		return
	}
	nextRule := s.Rules[s.d]
	for _, alt := range nextRule.Alts {
		//if alt.termSet[pset.termAlt] || alt == pset.termAlt {
		if child, isNew := pset.add(alt, s); isNew {
			pset.predictNext(child)
		}
		//}
	}
	//if nextRule.nullable {
	//	pset.predict(s.advance(nil))
	//}
}
