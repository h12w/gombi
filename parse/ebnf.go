package parse

import "strings"

//const noID = -42

var (
	EOF  = Term("EOF")
	Null = Term("Null")
	Self = Term("Self")
)

func newR() *R {
	return &R{}
}

func Term(name string) *R {
	return newR().As(name)
}

func Rule(name string, rules ...*R) *R {
	r := Con(rules...)
	r.initRecursiveRule(make(map[*R]bool), r)
	return r.As(name)
}
func (r *R) initRecursiveRule(m map[*R]bool, selfValue *R) {
	if m[r] {
		return
	}
	m[r] = true
	for _, alt := range r.Alts {
		if alt.Parent == Self {
			alt.Parent = selfValue
		}
		for i, cr := range alt.Rules {
			if cr == Self {
				alt.Rules[i] = selfValue
			} else {
				cr.initRecursiveRule(m, selfValue)
			}
		}
	}
}

func Con(rules ...*R) *R {
	if len(rules) == 1 {
		return rules[0]
	}
	r := newR()
	r.Alts = Alts{{r, rules}}
	return r
}

func Or(rules ...*R) *R {
	if len(rules) == 1 {
		return rules[0]
	}
	r := newR()
	r.Alts = make(Alts, len(rules))
	for i := range rules {
		r.Alts[i] = rules[i].toAlt(r)
	}
	return r
}
func (r *R) toAlt(parent *R) *Alt {
	if len(r.Alts) == 1 && r.name == "" {
		return &Alt{parent, r.Alts[0].Rules}
	}
	return &Alt{parent, Rules{r}}
}

func (r *R) As(name string) *R {
	r.name = name
	return r
}

func (r *R) ZeroOrOne() *R {
	return Or(r, Null).As(parens(r.Name()) + "?")
}

func (r *R) OneOrMore() *R {
	return Con(r, r.ZeroOrMore()).As(parens(r.Name()) + "+")
}

func (r *R) ZeroOrMore() *R {
	x := newR().As(parens(r.Name()) + "*")
	x.Alts = Or(Con(r, x), Null).toAlts(x)
	return x
}
func (r *R) toAlts(parent *R) Alts {
	r.eachAlt(func(a *Alt) {
		a.Parent = parent
	})
	return r.Alts
}

func parens(s string) string {
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return s
	}
	return "(" + s + ")"
}
