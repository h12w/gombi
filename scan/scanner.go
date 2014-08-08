package scan

import (
	"fmt"
	"io"
	"strings"
)

type Scanner struct {
	rx  rx
	buf *Buffer
	err error
	tok Token
}

type Token struct {
	Type  int
	Pos   int
	Value []byte
}

func NewScanner(pat string, r io.Reader) (*Scanner, error) {
	return &Scanner{buf: NewBuffer(r), rx: rx{compile(pat)}}, nil
}

func NewStringScanner(pat, text string) (*Scanner, error) {
	return &Scanner{buf: NewBuffer(strings.NewReader(text)), rx: rx{compile(pat)}}, nil
}

func (s *Scanner) Scan() bool {
	m := s.rx.FindReaderSubmatchIndex(s.buf)
	if m == nil {
		if _, _, err := s.buf.ReadRune(); err != nil {
			s.err = io.EOF
		} else {
			s.err = fmt.Errorf("token patterns do not cover all possible input, %s, %s.", string(s.buf.bytes()), s.rx.String())
		}
		return false
	}
	for i := 2; i < len(m)-1; i += 2 {
		if m[i] != -1 {
			token, pos, err := s.buf.ReadToken(m[i+1])
			if err != nil {
				s.err = err
				return false
			}
			s.tok = Token{
				Type:  i/2 - 1,
				Value: append([]byte{}, token...),
				Pos:   pos,
			}
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
