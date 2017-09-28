package scan

import (
	"errors"
	"io"

	"h12.me/dfa"
)

var invalidInputErr = errors.New("invalid input")

type Scanner struct {
	*Matcher

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
		matched    bool
		pos        = s.p
		matchedPos = s.p
		cur        = s.s0
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
		return true
	} else if s.p == len(s.src) {
		s.tok.ID = s.EOF
		s.tok.Lo = s.p
		s.tok.Hi = s.p
		s.err = io.EOF
		return true
	}
	s.tok.ID = s.Illegal
	s.tok.Lo = s.p
	s.tok.Hi = pos // record the error position
	s.err = invalidInputErr
	s.p++ // advance 1 byte when illegal
	return true
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

func (t *Token) Copy() *Token {
	c := *t
	return &c
}
