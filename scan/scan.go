package scan

import (
	"errors"
	"io"

	"github.com/hailiang/dfa"
)

var invalidInputErr = errors.New("invalid input")

type Scanner struct {
	*Matcher

	s       *dfa.FastS // current state
	s0      *dfa.FastS // start state cache
	src     []byte     // source buffer
	p       int        // byte position in buffer
	tp      int        // token start position in buffer
	mp      int        // matched position
	matched bool       // matched

	tok Token
	err error
}
type Token struct {
	ID int
	Lo int
	Hi int
}

func (s *Scanner) SetSource(src []byte) {
	s.s = &s.fast.States[0]
	s.s0 = &s.fast.States[0]
	s.src = src
	s.p = 0
	s.tp = 0
	s.mp = 0
	s.matched = false
	s.tok = Token{}
	s.err = nil
}

func (s *Scanner) SetPos(p int) {
	s.p = p
	s.tp = p
	s.mp = p
	s.matched = false
}

func (s *Scanner) Scan() bool {
	var (
		pos        = s.p
		cur        = s.s
		matched    = s.matched
		matchedPos = s.mp
	)
	for {
		if cur.Label >= 0 {
			s.tok.ID = cur.Label
			matchedPos = pos
			matched = true
		}
		if pos == len(s.src) {
			break
		}
		b := s.src[pos]
		if cur = cur.Trans[b]; cur == nil {
			break
		}
		pos++
	}
	if matched {
		s.mp = matchedPos
		s.matched = false
		s.tok.Lo = s.tp
		s.tok.Hi = matchedPos
		s.tp = matchedPos
		s.p = s.tp
		s.s = s.s0
		return true
	} else if pos == len(s.src) {
		s.tok.ID = s.eof
		s.tok.Lo = s.tp
		s.tok.Hi = s.tp
		s.err = io.EOF
		return false
	}
	s.tok.Lo = s.tp
	s.tok.Hi = s.tp + 1 // advance 1 byte when illegal
	s.tok.ID = s.illegal
	s.err = invalidInputErr
	s.tp++
	s.p = s.tp
	return false
}

func (s *Scanner) Pos() int {
	return s.p
}

func (s *Scanner) Token() *Token {
	return &s.tok
}

func (s *Scanner) Error() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}
