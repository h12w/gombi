package parse

type matchingRule struct {
	*Alt
	values []*state
	value  interface{}
	d      int // dot position
}

func (r *matchingRule) nextChildRule() *R {
	if r.d < len(r.Alt.Rules) {
		if r.Alt.Rules[r.d] == Null {
			r.d++ // skip trivial null rule
			return r.nextChildRule()
		}
		return r.Alt.Rules[r.d]
	}
	return nil
}

func (r *matchingRule) complete() bool {
	return r.d == len(r.Alt.Rules)
}

func (r *matchingRule) equal(o *state) bool {
	return r.Alt == o.Alt && r.d == o.d
}

func (r *matchingRule) expect(o *R) bool {
	return !r.complete() && r.nextChildRule() == o
}

type state struct {
	matchingRule
	parents *stateSet // multiple alternatives (parents) may share a common prefix (child)
}

func newState(alt *Alt) *state {
	return &state{
		matchingRule: matchingRule{
			Alt:    alt,
			values: make([]*state, len(alt.Rules)),
		},
		parents: newStateSet(nil),
	}
}

// newTermState intializes a parsed state for a terminal rule from a token.
func newTermState(t *Token) *state {
	return &state{
		matchingRule: matchingRule{
			Alt:   &Alt{Parent: t.R},
			d:     1,
			value: t.Value,
		},
		parents: newStateSet(nil)}
}

func (s *state) copy() *state {
	c := *s
	c.values = append([]*state{}, s.values...)
	return &c
}

func (s *state) step() *state {
	s.d++
	return s
}

// scan matches t with the expected input. If matched, it advances itself and
// returns true, otherwise, returns false.
func (s *state) scan(t *state) bool {
	if s.expect(t.rule()) {
		s.values[s.d] = t
		s.step()
		return true
	}
	return false
}
func (a *Alt) rule() *R {
	return a.Parent
}

type stateSet struct {
	a []*state
}

func newStateSet(r *R) *stateSet {
	ss := &stateSet{}
	if r != nil {
		r.eachAlt(func(a *Alt) {
			ss.add(newState(a), nil)
		})
	}
	return ss
}

func (ss *stateSet) add(o, parent *state) (isNew bool) {
	if s, ok := ss.find(o); ok {
		o = s
	} else {
		ss.a = append(ss.a, o)
		isNew = true
	}
	if parent != nil {
		o.parents.add(parent, nil)
	}
	return
}
func (ss *stateSet) find(o *state) (*state, bool) {
	for _, s := range ss.a {
		if s.equal(o) {
			return s, true
		}
	}
	return nil, false
}

func (ss *stateSet) each(visit func(*state)) {
	for _, s := range ss.a {
		visit(s)
	}
}
