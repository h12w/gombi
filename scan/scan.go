package scan

import (
	"fmt"
	"io"
	"strconv"
)

type Scanner struct {
	Matcher *TokenMatcher

	src []byte
	p   int

	tok Token
	err error
}
type Token struct {
	ID    int
	Value []byte
	Pos   int
}

func (s *Scanner) Scan() bool {
	if s.err == io.EOF {
		return false
	}

	buf := s.src[s.p:]
	id, size := s.Matcher.Match(buf)
	s.tok.ID = id
	s.tok.Value = buf[:size]
	s.tok.Pos = s.p
	s.p += size

	switch id {
	case s.Matcher.EOF:
		s.err = io.EOF
	case s.Matcher.Illegal:
		s.err = invalidInputError(buf)
		return false
	}
	return true
}
func invalidInputError(buf []byte) error {
	return fmt.Errorf("token pattern does not match input from %s.", strconv.Quote(string(prefix(buf, 20))))
}
func prefix(buf []byte, i int) []byte {
	if i <= len(buf) {
		return buf[:i]
	}
	return buf
}

func (s *Scanner) SetSource(src []byte) {
	s.src = src
	s.p = 0
	s.tok = Token{}
	s.err = nil
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
