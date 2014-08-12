package parse

func Con(rules ...*R) *R {
	return &R{"", Alts{{rules}}}
}

func Or(rules ...*R) *R {
	alts := make(Alts, len(rules))
	for i := range rules {
		alts[i] = &Alt{rules[i].toRules()}
	}
	return &R{"", alts}
}

func Rule(name string) *R {
	return &R{Name: name}
}

func (r *R) Con(rules ...*R) *R {
	r.Alts = Con(rules...).Alts
	return r
}

func (r *R) Or(rules ...*R) *R {
	r.Alts = Or(rules...).Alts
	return r
}

func (r *R) toRules() Rs {
	if len(r.Alts) == 1 {
		return r.Alts[0].Rs
	}
	return Rs{r}
}
