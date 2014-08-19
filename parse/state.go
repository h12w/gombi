package parse

type state struct {
	*matchingRule
	values []*state
	d      int // dot position
}

// matchingRule is the unchanged part during scanning.
// multiple parents are used to handle children with a common prefix or left
// recursive rule like: S ::= S + M.
type matchingRule struct {
	*Alt
	parents []*state
	token   *Token
}
type Token struct {
	Value []byte
	Pos   int
}

func newState(alt *Alt) *state {
	return &state{
		matchingRule: &matchingRule{Alt: alt},
		values:       make([]*state, len(alt.Rules)),
	}
}

// newTermState intializes a parsed state for a terminal rule from a token.
func newTermState(t *Token, r *R) *state {
	s := newState(r.Alts[0])
	s.token = t
	return s
}

func (s *state) copy() *state {
	c := *s
	c.values = append([]*state(nil), s.values...)
	return &c
}

func (s *state) step() *state {
	s.d++
	return s
}

func (s *state) nextChildRule() *R {
	if s.Alt.Rules[s.d] == Null {
		s.step() // skip trivial null rule
	}
	return s.Alt.Rules[s.d]
}

func (s *state) complete() bool {
	return s.d == len(s.Alt.Rules)
}

// not copied because a terminal state can never be a parent of multiple children
func (s *state) scan(t *Token, r *R) bool {
	if s.rule() == r {
		s.token = t
		s.step()
		return true
	}
	return false
}

// copied because multiple alternatives shares the same parent
func (s *state) advance(t *state) *state {
	ns := s.copy()
	ns.values[s.d] = t
	return ns.step()
}

type states struct {
	a []*state
}

func (ss *states) reset() states {
	ss.a = ss.a[:0]
	return *ss
}

func (ss *states) append(s *state) {
	ss.a = append(ss.a, s)
}

func (ss *states) each(visit func(*state)) {
	for _, s := range ss.a {
		visit(s)
	}
}

type stateSet struct {
	states
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
