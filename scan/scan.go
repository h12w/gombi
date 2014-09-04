package scan

import (
	"fmt"
	"io"
	"strconv"

	"github.com/hailiang/dfa"
)

type Scanner struct {
	*Matcher

	src []byte
	p   int
	s0  *dfa.FastS

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

	s.tok.Pos = s.p
	if s.p >= len(s.src) {
		s.tok.ID = s.eof
		s.tok.Value = nil
		s.err = io.EOF
		return false
	}
	{
		var (
			cur        = s.s0
			pos        = s.p
			matched    = false
			matchedPos = pos
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
			if cur = cur.Trans[s.src[pos]]; cur == nil {
				break
			}
			pos++
		}
		if matched {
			size := matchedPos - s.p
			s.tok.Value = s.src[s.p:matchedPos]
			s.p += size
			return true
		}
	}
	s.tok.ID = s.illegal
	s.tok.Value = s.src[s.p : s.p+1] // advance 1 byte when illegal
	s.err = invalidInputError(s.src[s.p:])
	s.p++
	return false
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
	s.s0 = &s.fast.States[0]
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
