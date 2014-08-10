package main

import "fmt"

func main() {
	P := grammar()
	s0 := newStateSet()
	s0.add(&state{P.id, P.term, P.alts[0], 0, 0})
	ss := []*stateSet{s0}
	tokens := []string{"2", "+", "3", "*", "4"}

	for i := 0; i <= len(tokens); i++ {
		ss = append(ss, newStateSet())
		ss[i].each(func(s state) {
			if s.complete() { // complete
				ss[s.i].each(func(os state) {
					if os.nextIs(s.id) {
						os.d++
						ss[i].add(&os)
					}
				})
			} else if next := s.next(); next.term { // scan
				if i < len(tokens) && s.nextIs(tokens[i]) {
					s.d++
					ss[i+1].add(&s)
				}
			} else { // predict
				for _, alt := range next.alts {
					ss[i].add(&state{next.id, next.term, alt, i, 0})
				}
			}
		})
		fmt.Println("r:", ss[i].expr())
	}
}

func grammar() *rule {
	T := &rule{id: "T", alts: alts{
		{rules{term("1")}},
		{rules{term("2")}},
		{rules{term("3")}},
		{rules{term("4")}},
	}}

	M := &rule{id: "M"}
	M.alts = alts{
		{rules{M, term("*"), T}},
		{rules{T}},
	}

	S := &rule{id: "S"}
	S.alts = alts{
		{rules{S, term("+"), M}},
		{rules{M}},
	}

	P := &rule{id: "P", alts: alts{{rules{S}}}}
	return P
}
