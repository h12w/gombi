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
	}
	for _, a := range r.Alts {
		if a.last() != EOF {
			a.Rules = append(a.Rules, EOF)
		}
	}
	return r
}

func (p *Parser) Parse(t *Token, tr *R) bool {
	p.s = p.predict(tr)
	if p.s == nil {
		return false
	}
	p.s.scan(t)
	p.propagate(p.s)
	return true
}

func (p *Parser) predict(tr *R) *state {
	pset := newStateSet(tr)
	if p.s == nil {
		p.r.eachAlt(func(alt *Alt) {
			pset.predict(newState(alt))
		})
	} else {
		pset.predict(p.s)
	}
	//fmt.Printf("### term state ->\n%s\n", pset.termState.dumpUp(0))
	//fmt.Println()
	//fmt.Printf("### predict set ->\n%s\n", pset.String())
	//fmt.Println()
	return pset.termState
}

func (p *Parser) Error() error {
	return nil
}

func (p *Parser) Results() []*Node {
	return p.results
}

func (p *Parser) propagate(s *state) {
	if s.complete() {
		if s.last().isEOF() {
			p.results = append(p.results, newNode(s))
		} else {
			for _, parent := range s.parents {
				p.propagate(parent.advance(s))
			}
		}
	}
	return
}

func (pset *stateSet) predict(s *state) {
	if s.complete() {
		for _, parent := range s.parents {
			pset.predict(parent.advance(s))
		}
		return
	}
	if s.isTerm {
		return
	}
	s.nextChildRule().eachAlt(func(alt *Alt) {
		if alt.isNull() {
			// copied because other alternatives should not be skipped
			pset.predict(s.copy().step())
		} else if child, isNew := pset.add(alt, s); isNew {
			pset.predict(child)
		}
	})
}
