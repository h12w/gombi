package parse

type (
	// R is a BNF production rule
	R struct {
		name string
		//		ID   int
		Alts
		recursive bool
	}
	Alt struct {
		*R
		Rules
	}
	Rules []*R
	Alts  []*Alt
)

func (r *R) isEOF() bool {
	return r == EOF
}

func (r *R) isNull() bool {
	return r == Null
}

func (r *R) isTerm() bool {
	return len(r.Alts) == 1 && len(r.Alts[0].Rules) == 1 && r.Alts[0].Rules[0] == nil
}

func (r *R) eachAlt(visit func(a *Alt)) {
	for _, a := range r.Alts {
		visit(a)
	}
}

func (r *R) traverseRule(m map[*R]bool, visit func(*R)) {
	if m[r] {
		return
	}
	m[r] = true
	visit(r)
	for _, a := range r.Alts {
		for _, rule := range a.Rules {
			rule.traverseRule(m, visit)
		}
	}
}

func (a *Alt) last() *R {
	if len(a.Rules) > 0 {
		return a.Rules[len(a.Rules)-1]
	}
	return nil
}

func (a *Alt) isNull() bool {
	return len(a.Rules) == 1 && a.Rules[0].isNull()
}
