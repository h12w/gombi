package parse

type token struct {
	value string
	*R
}

type matchingR struct {
	*R
	*Alt
	d int // dot position
}

func (r *matchingR) nextChildR() *R {
	return r.Alt.Rs[r.d]
}

func (r *matchingR) complete() bool {
	return r.d == len(r.Alt.Rs)
}

func (r *matchingR) equal(o *state) bool {
	return r.Alt == o.Alt && r.d == o.d
}

func (r *matchingR) expect(o *R) bool {
	return !r.complete() && r.nextChildR() == o
}

type state struct {
	matchingR
	parents stateSet // multiple alternatives (parents) may share a common prefix (child)
	values  []*state
	value   *string
}

func newState(r *R, alt *Alt) *state {
	return &state{
		matchingR: matchingR{
			R:   r,
			Alt: alt,
		},
		parents: *newStateSet(),
		values:  make([]*state, len(alt.Rs))}
}

func newTermState(t *token) *state {
	return &state{
		matchingR: matchingR{
			R: t.R,
		},
		parents: *newStateSet(),
		value:   &t.value}
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
