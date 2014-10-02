package parse

type Parser struct {
	r       *R
	s       *state
	results []*Node
}

func New(r *R) *Parser {
	r.appendEOF()
	p := &Parser{r: r}
	p.Reset()
	return p
}
func (p *Parser) Reset() {
	p.results = nil
	p.s = nil
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

func (p *Parser) Parse(t *Token, tr *R) bool {
	pset := &stateSet{termRule: tr}
	if p.s == nil {
		p.r.eachAlt(func(alt *Alt) {
			pset.predictAll(newState(alt), tr)
		})
	} else {
		pset.predictAll(p.s, tr)
	}
	p.s = pset.termState
	if p.s == nil {
		return false
	}
	if p.s.scan(t, tr) {
		p.results = append(p.results, p.s.propagate()...)
	}
	//fmt.Printf("### term state ->\n%s\n", p.s.dumpUp(0))
	//fmt.Println()
	//fmt.Printf("### predict set ->\n%s\n", pset.String())
	//fmt.Println()
	return true
}

func (p *Parser) Error() error {
	return nil
}

func (p *Parser) Results() (rs []*Node) {
	return p.results
}

func (pset *stateSet) predictAll(s *state, termRule *R) {
	if s.complete() {
		for _, parent := range s.parents {
			pset.predictAll(parent.advance(s), termRule)
		}
	} else {
		pset.predict(s, termRule)
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

func (pset *stateSet) predict(s *state, termRule *R) {
	if s.isTerm {
		return
	}
	s.nextChildRule().eachAlt(func(alt *Alt) {
		if alt.isNull() {
			// copied because other alternatives should not be skipped
			pset.predictNull(s.copy().step(), termRule)
		} else if child, isNew := pset.add(alt, s); isNew {
			pset.predict(child, termRule)
		}
	})
}

func (pset *stateSet) predictNull(s *state, termRule *R) {
	if s.complete() {
		for _, parent := range s.parents {
			pset.predictNull(parent.advance(s), termRule)
		}
	} else {
		pset.predict(s, termRule)
	}
}
