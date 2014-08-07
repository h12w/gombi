package scan

import (
	"fmt"
	"io"
	"io/ioutil"
)

type Scanner struct {
	rx
	buf []byte
	pos int
	err error
	tok Token
}

type Token struct {
	Type  int
	Pos   int
	Value []byte
}

func NewScanner(pat string, r io.Reader) (*Scanner, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &Scanner{buf: buf, rx: rx{compile(pat)}}, nil
}

func NewBufferScanner(pat string, buf []byte) (*Scanner, error) {
	return &Scanner{buf: buf, rx: rx{compile(pat)}}, nil
}

func (s *Scanner) Scan() bool {
	if s.pos >= len(s.buf) {
		s.err = io.EOF
		return false
	}
	m := s.FindSubmatchIndex(s.buf[s.pos:])
	if m == nil {
		s.err = fmt.Errorf("token patterns do not cover all possible input, %s, %s.", string(s.buf[s.pos:]), s.Regexp.String())
		return false
	}
	for i := 2; i < len(m)-1; i += 2 {
		if m[i] != -1 {
			s.tok = Token{
				Type:  i/2 - 1,
				Value: s.buf[s.pos:][0:m[i+1]],
				Pos:   s.pos,
			}
			s.pos += m[i+1]
			return true
		}
	}
	return false
}

func (s *Scanner) Token() Token {
	return s.tok
}

func (s *Scanner) Error() error {
	return s.err
}
