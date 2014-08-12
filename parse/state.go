package parse

type matchingRule struct {
	*R
	*Alt
	d int // dot position
}

func (r *matchingRule) nextChildRule() *R {
	if r.d < len(r.Alt.Rs) {
		return r.Alt.Rs[r.d]
	}
	return nil
}

func (r *matchingRule) nextIsNullable() (result bool) {
	next := r.nextChildRule()
	next.eachAlt(func(r *R, alt *Alt) {
		if len(alt.Rs) == 1 && alt.Rs[0] == Null {
			result = true
		}
	})
	return result
}

func (r *matchingRule) complete() bool {
	return r.d == len(r.Alt.Rs)
}

func (r *matchingRule) equal(o *state) bool {
	return r.Alt == o.Alt && r.d == o.d
}

func (r *matchingRule) expect(o *R) bool {
	return !r.complete() && r.nextChildRule() == o
}

type state struct {
	matchingRule
	parents stateSet // multiple alternatives (parents) may share a common prefix (child)
	values  []*state
	value   interface{}
}

func newState(r *R, alt *Alt) *state {
	return &state{
		matchingRule: matchingRule{
			R:   r,
			Alt: alt,
		},
		parents: *newStateSet(),
		values:  make([]*state, len(alt.Rs))}
}

func newTermState(t *Token) *state {
	return &state{
		matchingRule: matchingRule{
			R: t.R,
		},
		parents: *newStateSet(),
		value:   t.Value}
}

func (s *state) copy() *state {
	c := *s
	c.values = append([]*state{}, s.values...)
	return &c
}

// scan matches t with the expected input. If matched, it advances itself and
// returns true, otherwise, returns false.
func (s *state) scan(t *state) bool {
	if s.expect(t.R) {
		s.values[s.d] = t
		s.d++
		return true
	}
	return false
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
