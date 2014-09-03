package scan

import (
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
)

type Scanner struct {
	*Matcher

	src []byte
	p   int

	toks []Token
	tok  *Token
	err  error
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
	id, size := s.Matcher.matchBytes(buf)
	switch id {
	case s.EOF:
		if s.p == len(s.src) {
			s.reachEOF(s.p)
			return true
		}
	case s.Illegal:
		s.err = invalidInputError(buf)
		s.tok = &Token{ID: s.Illegal, Value: buf[:1], Pos: s.p}
		s.p++ // advance 1 byte to avoid indefinate loop
		return false
	}
	s.tok = &Token{
		ID:    id,
		Value: buf[:size],
		Pos:   s.p,
	}
	s.p += size
	return true
}
func (s *Scanner) reachEOF(pos int) {
	s.err = io.EOF
	s.tok = &Token{
		ID:    s.EOF,
		Value: nil,
		Pos:   pos,
	}
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
	s.reset()
	s.src = src
}

func (s *Scanner) SetReader(r io.Reader) error {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	s.SetSource(src)
	return nil
}

func (s *Scanner) reset() {
	s.p = 0
	s.err = nil
	s.tok = nil
}

func (s *Scanner) Pos() int {
	return s.p
}

func (s *Scanner) Token() *Token {
	return s.tok
}

func (s *Scanner) Error() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}
