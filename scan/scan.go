package scan

import (
	"fmt"
	"io"
	"io/ioutil"
)

type ByteScanner struct {
	rx  rx
	buf []byte
	p   int
	err error
	tok Token
}

func NewByteScanner(pat string, r io.Reader) (*ByteScanner, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &ByteScanner{rx: rx{compile(pat)}, buf: buf}, nil
}

func (s *ByteScanner) Scan() bool {
	buf := s.buf[s.p:]
	m := s.rx.FindSubmatchIndex(buf)
	if m == nil {
		if s.p == len(s.buf) {
			s.err = io.EOF
		} else {
			s.err = fmt.Errorf("token patterns do not cover all possible input, %s, %s.", string(buf), s.rx.String())
		}
		return false
	}
	for i := 2; i < len(m)-1; i += 2 {
		if m[i] != -1 {
			s.tok = Token{
				Type:  i/2 - 1,
				Value: append([]byte(nil), buf[:m[i+1]]...),
				Pos:   s.p,
			}
			s.p += m[i+1]
			return true
		}
	}
	return false
}

func (s *ByteScanner) Token() Token {
	return s.tok
}

func (s *ByteScanner) Error() error {
	return s.err
}

type UTF8Scanner struct {
	rx  rx
	buf *runeBuffer
	err error
	tok Token
}

type Token struct {
	Type  int
	Pos   int
	Value []byte
}

func NewUTF8Scanner(pat string, r io.Reader) (*UTF8Scanner, error) {
	return &UTF8Scanner{buf: newRuneBuffer(r), rx: rx{compile(pat)}}, nil
}

func (s *UTF8Scanner) Scan() bool {
	m := s.rx.FindReaderSubmatchIndex(s.buf)
	if m == nil {
		if _, _, err := s.buf.ReadRune(); err != nil {
			s.err = err
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

func (s *UTF8Scanner) Token() Token {
	return s.tok
}

func (s *UTF8Scanner) Error() error {
	return s.err
}
