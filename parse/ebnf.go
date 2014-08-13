package parse

func Term(name string) *R {
	return &R{Name: name}
}

func Rule(name string, rules ...*R) *R {
	r := Con(rules...).As(name)
	r.traverseAlt(make(map[*R]bool), func(a *Alt) {
		for i := range a.Rs {
			if a.Rs[i] == Self {
				a.Rs[i] = r
			}
		}
	})
	return r
}

func Con(rules ...*R) *R {
	if len(rules) == 1 && rules[0].Name == "" {
		return rules[0]
	}
	return &R{"", Alts{{rules}}}
}

func Or(rules ...*R) *R {
	if len(rules) == 1 && rules[0].Name == "" {
		return rules[0]
	}
	alts := make(Alts, len(rules))
	for i := range rules {
		alts[i] = rules[i].toAlt()
	}
	return &R{"", alts}
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
	x.Alts = Con(x, r).ZeroOrOne().Alts
	return x
}

func (r *R) toAlt() *Alt {
	if len(r.Alts) == 1 {
		return &Alt{r.Alts[0].Rs}
	}
	return &Alt{Rs{r}}
}

func (r *R) toAlts() Alts {
	return r.Alts
}
