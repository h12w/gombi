package parse

type token struct {
	value string
	*rule
}

type state struct {
	*rule
	*alt
	d       int      // dot position
	parents stateSet // multiple alternatives (parents) may share a common prefix (child)
	values  []*state
	value   *string
}

func newState(r *rule, i int) *state {
	return &state{
		rule:    r,
		alt:     r.alts[i],
		parents: *newStateSet(),
		values:  make([]*state, len(r.alts[i].rules))}
}

func newTermState(t *token) *state {
	return &state{
		rule:    t.rule,
		parents: *newStateSet(),
		value:   &t.value}
}

func (s *state) addParent(parent *state) {
	s.parents.add(parent)
}

func (s *state) copy() *state {
	c := *s
	c.values = append([]*state{}, s.values...)
	return &c
}

func (s *state) equal(o *state) bool {
	return s.alt == o.alt && s.d == o.d
}

func (s *state) complete() bool {
	return s.d == len(s.alt.rules)
}

func (s *state) next() *rule {
	return s.alt.rules[s.d]
}

// scan matches t with the expected input. If matched, it advances itself and
// returns true, otherwise, returns false.
func (s *state) scan(t *state) bool {
	if s.expect(t.rule) {
		s.values[s.d] = t
		s.d++
		return true
	}
	return false
}

func (s *state) expect(r *rule) bool {
	return !s.complete() && s.alt.rules[s.d] == r
}

type stateSet struct {
	a []*state
}

func newStateSet() *stateSet {
	return &stateSet{}
}

func (ss *stateSet) empty() bool {
	return len(ss.a) == 0
}

func (ss *stateSet) add(o *state) (*state, bool) {
	for _, s := range ss.a {
		if s.equal(o) {
			return s, false
		}
	}
	ss.a = append(ss.a, o)
	return o, true
}

func (ss *stateSet) each(visit func(*state)) {
	for _, s := range ss.a {
		visit(s)
	}
}
