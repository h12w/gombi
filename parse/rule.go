package parse

type (
	// R is a BNF production rule
	R struct {
		Name string
		Alts
	}
	Alt struct {
		Rs
	}
	Rs   []*R
	Alts []*Alt
)

var (
	EOF  = Rule("EOF")
	Null = Rule("Null")
)

func (r *R) isEOF() bool {
	return r == EOF
}

func (r *R) isNull() bool {
	return r == Null
}

func (r *R) eachAlt(visit func(r *R, a *Alt)) {
	if r == nil {
		return
	}
	for _, a := range r.Alts {
		visit(r, a)
	}
}

func (a Alt) last() *R {
	if len(a.Rs) > 0 {
		return a.Rs[len(a.Rs)-1]
	}
	return nil
}

func (a Alt) isNull() bool {
	return len(a.Rs) == 1 && a.Rs[0].isNull()
}

func (r *R) appendEOF() {
	for _, a := range r.Alts {
		if a.last() != EOF {
			a.Rs = append(a.Rs, EOF)
		}
	}
}
