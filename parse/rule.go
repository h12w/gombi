package parse

/*
Why BNF?
1. BNF is the minimal representation of context free grammars.
2. EBNF can be represented with BNF.
3. Loop is represented with recursive rule in BNF, while in EBNF it can be
   directly defined. However, recursive rule is still needed to represent
   grammars like infix expressions.

Rule naming
1. Nested rules are reduced as much as possible, but named rules are always kept.
2. An unnamed rule has a display name. If it is not recursive, its display name
   is its definition wrapped by parenthesis, otherwise, its display name is
   generated from its pointer address.
3. Rules generated by ZeroOrMore and OneOrMore are automatically named as r* and
   r+.
*/

type (
	// R is a BNF production rule
	R struct {
		name string
		Alts
	}
	Alt struct {
		*R
		Rules
		termSet altSet
	}
	Rules   []*R
	Alts    []*Alt
	Builder struct {
		terms map[string]*R
	}
	altSet map[*Alt]bool
)

var (
	EOF = newTerm().As("EOF")
	SOF = newTerm().As("SOF")
)

func NewBuilder() *Builder {
	return &Builder{terms: make(map[string]*R)}
}

func (r *R) InitTermSet() {
	m := make(altSet)
	r.eachAlt(func(alt *Alt) {
		alt.initTermSet(m)
	})
}

func (a *Alt) initTermSet(m map[*Alt]bool) altSet {
	if m[a] {
		return a.termSet
	}
	m[a] = true
	if len(a.Rules) > 0 {
		r0 := a.Rules[0]
		if len(r0.Alts) == 1 {
			a.termSet = r0.Alts[0].initTermSet(m)
		} else {
			r0.eachAlt(func(alt *Alt) {
				for aa := range alt.initTermSet(m) {
					a.termSet[aa] = true
				}
			})
		}
		for i := 1; i < len(a.Rules); i++ {
			a.Rules[i].eachAlt(func(alt *Alt) {
				alt.initTermSet(m)
			})
		}
	}
	return a.termSet
}

func newAlt(parent *R, rules Rules) *Alt {
	a := &Alt{parent, rules, make(altSet)}
	//alt := &Alt{parent, rules, make(altSet)}
	//alt.initTermSet()
	return a
}

//func (a *Alt) initTermSet() {
//	if len(a.Rules) > 0 {
//		for _, alt := range a.Rules[0].Alts {
//			for t := range alt.termSet {
//				a.termSet[t] = true
//			}
//		}
//	}
//}

func newTerm() *R {
	r := &R{}
	alt := newAlt(r, nil)
	alt.termSet = map[*Alt]bool{alt: true}
	r.Alts = Alts{alt}
	return r
}

func (r *R) isTerm() bool {
	return len(r.Alts) == 1 && len(r.Alts[0].Rules) == 0
}

func NewRule() *R {
	return &R{}
}

func (b *Builder) Term(name string) *R {
	if r, ok := b.terms[name]; ok {
		return r
	}
	r := newTerm()
	r = r.As(name)
	b.terms[name] = r
	return r
}

func (b *Builder) toRules(a []interface{}) []*R {
	rs := make([]*R, len(a))
	for i := range rs {
		switch o := a[i].(type) {
		case *R:
			rs[i] = o
		case string:
			rs[i] = b.Term(o)
		default:
			panic("type of argument should be either string or *R")
		}
	}
	return rs
}

func (r *R) Define(o *R) *R {
	r.Alts = o.Alts
	for i := range r.Alts {
		r.Alts[i].R = r
		//r.Alts[i].initTermSet()
	}
	return r
}

func (b *Builder) Con(rs ...interface{}) *R {
	return con(b.toRules(rs)...)
}

func con(rules ...*R) *R {
	if len(rules) == 1 {
		return rules[0]
	}
	r := NewRule()
	r.Alts = Alts{newAlt(r, rules)}
	return r
}

func (b *Builder) Or(rs ...interface{}) *R {
	return or(b.toRules(rs)...)
}
func or(rules ...*R) *R {
	if len(rules) == 1 {
		return rules[0]
	}
	r := NewRule()
	r.Alts = make(Alts, len(rules))
	for i := range rules {
		r.Alts[i] = rules[i].toAlt(r)
	}
	return r
}
func (r *R) toAlt(parent *R) *Alt {
	if len(r.Alts) == 1 && r.name == "" { // reduce unnamed rule
		return newAlt(parent, r.Alts[0].Rules)
	}
	return newAlt(parent, Rules{r})
}

func (r *R) As(name string) *R {
	r.name = name
	return r
}

func (r *R) AtLeast(n int) *R {
	if n == 1 {
		return r.oneOrMore()
	} else if n > 1 {
		rs := make(Rules, n)
		for i := range rs {
			rs[i] = r
		}
		rs[n-1] = r.oneOrMore()
		return con(rs...)
	}
	panic("n should be at larger than 0")
}

func (r *R) Repeat(limit ...int) *R {
	switch len(limit) {
	//	case 0:
	//		return r.zeroOrMore()
	case 1:
		n := limit[0]
		rs := make(Rules, n)
		for i := range rs {
			rs[i] = r
		}
		return con(rs...)
	case 2:
		lo, hi := limit[0], limit[1]
		if lo > hi {
			lo, hi = hi, lo
		}
		rs := make(Rules, 0, hi-lo+1)
		for n := lo; n <= hi; n++ {
			rs = append(rs, r.Repeat(n))
		}
		return or(rs...)
	}
	panic("repeat should have zero to two arguments")
}

func (r *R) oneOrMore() *R {
	x := NewRule()
	x.Define(or(r, con(r, x)))
	x.As(parens(r.Name()) + "+")
	return x
}

func (r *R) eachAlt(visit func(a *Alt)) {
	for _, a := range r.Alts {
		visit(a)
	}
}

func (a *Alt) rule() *R {
	return a.R
}
