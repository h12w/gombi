package parse

type (
	matchingRule struct {
		*Rule
		pos int
	}
	matchingRules []matchingRule
	state         struct {
		kernel matchingRules
		next   transTable
	}
	transTable []trans
	trans      struct {
		input *Rule
		next  *state
	}
)

func mergeTransTable(ts []transTable) (table transTable) {
	for _, t := range ts {
		for _, trans := range t {
			for i := range table {
				if table[i].input == trans.input {
				}
			}
		}
	}
	return
}

func (r *matchingRule) step() matchingRule {
	cr := *r
	cr.pos++
	return cr
}

func (r *matchingRule) next() (table transTable) {
	switch r.typ {
	case conRule:
		if r.pos < len(r.Rules) {
			table = append(table, trans{
				input: r.Rules[r.pos],
				next: &state{
					kernel: matchingRules{r.step()},
				},
			})
		}
	case orRule:
		if r.pos == 0 {
			for _, input := range r.Rules {
				table = append(table, trans{
					input: input,
					next: &state{
						kernel: matchingRules{r.step()},
					},
				})
			}
		}
	}
	return
}

func (rs matchingRules) closure() (result matchingRules) {
	set := make(map[matchingRule]bool)
	stack := make(matchingRules, 0, len(rs))
	for i := len(rs) - 1; i >= 0; i-- {
		stack = append(stack, rs[i])
	}
	var r matchingRule
	for len(stack) > 0 {
		stack, r = stack[:len(stack)-1], stack[len(stack)-1]
		if set[r] {
			continue
		} else {
			result = append(result, r)
			set[r] = true
		}
		switch r.typ {
		case conRule:
			if r.pos < len(r.Rules) {
				stack = append(stack, matchingRule{r.Rules[r.pos], 0})
			}
		case orRule:
			if r.pos == 0 {
				for i := len(r.Rules) - 1; i >= 0; i-- {
					stack = append(stack, matchingRule{r.Rules[i], 0})
				}
			}
		}
	}
	return
}

func (s matchingRules) Len() int {
	return len(s)
}
func (s matchingRules) Less(i, j int) bool {
	return s[i].id < s[j].id
}
func (s matchingRules) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
