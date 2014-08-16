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
	return &state{
		matchingRule: &matchingRule{
			Alt:   &Alt{Parent: r},
			token: t,
		},
		d: 1, // a term state is complete already.
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
	return s.d == len(s.Alt.Rules)
}

func (s *state) expect(o *R) bool {
	return !s.complete() && s.nextChildRule() == o
}

// scan matches t with the expected input. If matched, it records the value,
// advances itself and returns true, otherwise, returns false.
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
		ss.a = append(ss.a, newState(&Alt{Parent: newR(), Rules: Rules{r}}))
	}
	return ss
}

func (ss *stateSet) reset() *stateSet {
	ss.a = ss.a[:0]
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
