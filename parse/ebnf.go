package parse

func Term(name string) *R {
	return &R{Name: name}
}

func Rule(name string, rules ...*R) *R {
	return Con(rules...).initRecursiveRule().As(name)
}
func (r *R) initRecursiveRule() *R {
	r.traverseAlt(make(map[*R]bool), func(a *Alt) {
		if a.Parent == Self {
			a.Parent = r
		}
		for i := range a.Rules {
			if a.Rules[i] == Self {
				a.Rules[i] = r
			}
		}
	})
	return r
}

func Con(rules ...*R) *R {
	if len(rules) == 1 && rules[0].Name == "" {
		return rules[0]
	}
	r := &R{}
	r.Alts = Alts{{r, rules}}
	return r
}

func Or(rules ...*R) *R {
	if len(rules) == 1 && rules[0].Name == "" {
		return rules[0]
	}
	r := &R{Alts: make(Alts, len(rules))}
	for i := range rules {
		r.Alts[i] = rules[i].toAlt(r)
	}
	return r
}
func (r *R) toAlt(parent *R) *Alt {
	if len(r.Alts) == 1 {
		return &Alt{parent, r.Alts[0].Rules}
	}
	return &Alt{parent, Rules{r}}
}

func (r *R) As(name string) *R {
	r.Name = name
	return r
}

func (r *R) ZeroOrOne() *R {
	return Or(r, Null)
}

func (r *R) ZeroOrMore() *R {
	x := &R{}
	x.Alts = Con(x, r).ZeroOrOne().toAlts(x)
	return x
}
func (r *R) toAlts(parent *R) Alts {
	r.eachAlt(func(a *Alt) {
		a.Parent = parent
	})
	return r.Alts
}
