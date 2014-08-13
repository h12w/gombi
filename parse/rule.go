package parse

type (
	// R is a BNF production rule
	R struct {
		Name string
		Alts
	}
	Alt struct {
		Parent *R
		Rules
	}
	Rules []*R
	Alts  []*Alt
)

var (
	EOF  = Term("EOF")
	Null = Term("Null")
	Self = Term("Self")
)

func (r *R) isEOF() bool {
	return r == EOF
}

func (r *R) isNull() bool {
	return r == Null
}

func (r *R) eachAlt(visit func(a *Alt)) {
	for _, a := range r.Alts {
		visit(a)
	}
}

func (r *R) traverseAlt(m map[*R]bool, visit func(*Alt)) {
	m[r] = true
	for _, a := range r.Alts {
		visit(a)
		for _, rule := range a.Rules {
			if !m[r] {
				rule.traverseAlt(m, visit)
			}
		}
	}
}

func (a Alt) last() *R {
	if len(a.Rules) > 0 {
		return a.Rules[len(a.Rules)-1]
	}
	return nil
}

func (a Alt) isNull() bool {
	return len(a.Rules) == 1 && a.Rules[0].isNull()
}
