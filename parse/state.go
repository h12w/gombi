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
	ID    int
	Value []byte
	Pos   int
}

func newState(alt *Alt) *state {
	return &state{
		matchingRule: &matchingRule{Alt: alt},
		values:       make([]*state, len(alt.Rules)),
	}
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
	if s.isTerm {
		return s.d == 1
	}
	return s.d == len(s.Alt.Rules)
}

// not copied because a terminal state can never be a parent of multiple children
func (s *state) scan(t *Token, r *R) {
	s.token = t
	s.step()
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

type stateSet struct {
	termRule  *R
	termState *state
	states
}

func (ss *stateSet) add(o *Alt, parent *state) (child *state, isNew bool) {
	if s, ok := ss.find(o); ok {
		child = s
	} else {
		child = newState(o)
		ss.a = append(ss.a, child)
		isNew = true
	}
	if child.rule() == ss.termRule {
		ss.termState = child
	}
	// parent != nil
	child.parents = append(child.parents, parent)
	return
}
func (ss *stateSet) find(o *Alt) (*state, bool) {
	for _, s := range ss.a {
		if s.Alt == o {
			return s, true
		}
	}
	return nil, false
}
