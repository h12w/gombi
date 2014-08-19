package scan

import (
	"fmt"
	"io"
	"io/ioutil"
)

type Scanner interface {
	SetReader(r io.Reader) error
	SetMatcher(m *Matcher)
	SetEOF(id int)
	Scan() bool
	Token() *Token
	Error() error
	String() string
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

type ByteScanner struct {
	scannerBase
	buf []byte
	p   int
}

func NewByteScanner(m *Matcher) *ByteScanner {
	return &ByteScanner{scannerBase: scannerBase{matcher: m}}
}

func (s *ByteScanner) SetReader(r io.Reader) error {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	s.scannerBase.reset()
	s.buf = buf
	s.p = 0
	return nil
}

func (s *ByteScanner) Scan() bool {
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
		s.err = invalidInputError(s.matcher, buf)
		return false
	}
	s.tok = &Token{
		ID:    id,
		Value: append([]byte(nil), buf[:size]...),
		Pos:   s.p,
	}
	s.p += size
	return true
}

type UTF8Scanner struct {
	scannerBase
	buf *runeBuffer
}

func NewUTF8Scanner(m *Matcher) *UTF8Scanner {
	return &UTF8Scanner{scannerBase: scannerBase{matcher: m}}
}

func (s *UTF8Scanner) SetReader(r io.Reader) error {
	s.scannerBase.reset()
	s.buf = newRuneBuffer(r)
	return nil
}

func (s *UTF8Scanner) Scan() bool {
	if s.err == io.EOF {
		return false
	}
	id, size := s.matcher.matchReader(s.buf)
	if id == -1 {
		if _, _, err := s.buf.ReadRune(); err != nil {
			if err == io.EOF {
				s.reachEOF(s.buf.p)
				return true
			}
			s.err = err
			return false
		}
		s.err = invalidInputError(s.matcher, s.buf.bytes())
		return false
	}
	token, pos, err := s.buf.ReadToken(size)
	if err != nil {
		s.err = err
		return false
	}
	s.tok = &Token{
		ID:    id,
		Value: append([]byte{}, token...),
		Pos:   pos,
	}
	return true
}

func invalidInputError(m *Matcher, buf []byte) error {
	return fmt.Errorf("token pattern `%s` does not match input from `%s`.", m.String(), string(prefix(buf, 20)))
}

func prefix(buf []byte, i int) []byte {
	if i <= len(buf) {
		return buf[:i]
	}
	return buf
}
