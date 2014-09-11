package scan

import (
	"errors"
	"io"

	"github.com/hailiang/dfa"
)

var invalidInputErr = errors.New("invalid input")

type Scanner struct {
	*Matcher

	s   *dfa.FastS // current state
	s0  *dfa.FastS // start state cache
	src []byte     // source buffer
	p   int        // position in source buffer

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
	s.tok = Token{}
	s.err = nil
}

func (s *Scanner) SetPos(p int) {
	s.p = p
}

func (s *Scanner) Scan() bool {
	var (
		pos        = s.p
		cur        = s.s
		matchedPos = s.p
		matched    bool
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
		s.tok.Lo = s.p
		s.tok.Hi = matchedPos
		s.p = matchedPos
		s.s = s.s0
		return true
	} else if pos == len(s.src) {
		s.tok.ID = s.eof
		s.tok.Lo = s.p
		s.tok.Hi = s.p
		s.err = io.EOF
		return false
	}
	s.tok.Lo = s.p
	s.tok.Hi = s.p + 1 // advance 1 byte when illegal
	s.tok.ID = s.illegal
	s.err = invalidInputErr
	s.p++
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
