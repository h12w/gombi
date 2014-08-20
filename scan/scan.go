package scan

import (
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
)

type Scanner struct {
	scannerBase
	buf []byte
	p   int
}

func NewScanner(m *Matcher) *Scanner {
	return &Scanner{scannerBase: scannerBase{matcher: m}}
}

func (s *Scanner) Pos() int {
	return s.p
}

func (s *Scanner) SetBuffer(buf []byte) {
	s.scannerBase.reset()
	s.buf = buf
	s.p = 0
}

func (s *Scanner) SetReader(r io.Reader) error {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	s.scannerBase.reset()
	s.buf = buf
	s.p = 0
	return nil
}

func (s *Scanner) Scan() bool {
	if s.err == io.EOF {
		return false
	}
	buf := s.buf[s.p:]
	id, size := s.matcher.matchBytes(buf)
	if id == -1 {
		if s.p == len(s.buf) {
			s.reachEOF(s.p)
			return true
		}
		s.err = invalidInputError(buf)
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

type scannerBase struct {
	matcher *Matcher
	err     error
	tok     *Token
	eofID   int
}
type Token struct {
	ID    int
	Value []byte
	Pos   int
}

func (s *scannerBase) String() string {
	return s.matcher.String()
}

func (s *scannerBase) reset() {
	s.err = nil
	s.tok = nil
}

func (s *scannerBase) SetMatcher(m *Matcher) {
	s.matcher = m
}

func (s *scannerBase) Token() *Token {
	return s.tok
}

func (s *scannerBase) reachEOF(pos int) {
	s.err = io.EOF
	s.tok = &Token{
		ID:    s.eofID,
		Value: nil,
		Pos:   pos,
	}
}

func (s *scannerBase) SetEOF(id int) {
	s.eofID = id
}

func (s *scannerBase) Error() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
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
