package parse

func Rule(name string) *R {
	return &R{Name: name}
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

func (r *R) Is(rules ...*R) *R {
	r.Alts = Con(rules...).Alts
	//fmt.Println(rules, Con(rules...), r)
	return r
}

func (r *R) ZeroOrOne() *R {
	return Or(r, Null)
}

func (r *R) ZeroOrMore() *R {
	x := Rule("")
	x.Is(Con(x, r).ZeroOrOne())
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
