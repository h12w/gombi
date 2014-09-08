package scan

import "io"

type GotoScanner struct {
	Match        func(src []byte, pos int) (size, label int, matched bool)
	EOF, Illegal int

	src []byte
	p   int

	tok Token
	err error
}

func (s *GotoScanner) Scan() bool {
	if s.err == io.EOF {
		return false
	}

	s.tok.Pos = s.p
	if s.p >= len(s.src) {
		s.tok.ID = s.EOF
		s.tok.Value = nil
		s.err = io.EOF
		return false
	}
	if size, label, matched := s.Match(s.src, s.p); matched {
		s.tok.ID = label
		s.tok.Value = s.src[s.p : s.p+size]
		s.p += size
		return true
	}
	s.tok.ID = s.Illegal
	s.tok.Value = s.src[s.p : s.p+1] // advance 1 byte when illegal
	s.err = invalidInputErr
	s.p++
	return false
}

func (s *GotoScanner) SetSource(src []byte) {
	s.src = src
	s.p = 0
	s.tok = Token{}
	s.err = nil
}

func (s *GotoScanner) Pos() int {
	return s.p
}

func (s *GotoScanner) Token() *Token {
	return &s.tok
}

func (s *GotoScanner) Error() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}
