package parse

type state struct {
	*matchingRule
	values []*state
	d      int // dot position
}

// matchingRule is the unchanged part during scanning
type matchingRule struct {
	*Alt
	parents []*state // multiple alternatives (parents) may share a common prefix (child)
	value   interface{}
}

func newState(alt *Alt) *state {
	return &state{
		matchingRule: &matchingRule{
			Alt: alt,
		},
		values: make([]*state, len(alt.Rules)),
	}
}

// newTermState intializes a parsed state for a terminal rule from a token.
func newTermState(t *Token) *state {
	return &state{
		matchingRule: &matchingRule{
			Alt:   &Alt{Parent: t.R},
			value: t.Value,
		},
		d: 1,
	}
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

func (r *state) nextChildRule() *R {
	if r.Alt.Rules[r.d] == Null {
		r.step()
	}
	return r.Alt.Rules[r.d]
}

func (r *state) complete() bool {
	return r.d == len(r.Alt.Rules)
}

func (r *state) expect(o *R) bool {
	return !r.complete() && r.nextChildRule() == o
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
		o.parents = append(o.parents, parent)
	}
	return
}
func (ss *stateSet) find(o *state) (*state, bool) {
	for _, s := range ss.a {
		if s.Alt == o.Alt && s.d == o.d {
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
