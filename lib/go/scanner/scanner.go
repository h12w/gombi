package scanner

import (
	"fmt"
	"path/filepath"
	"unicode/utf8"

	"go/token"
)

const (
	firstOp = int(token.ADD)
	lastOp  = int(token.COLON)
)

const (
	ScanComments    Mode = 1 << iota // return comments as COMMENT tokens
	dontInsertSemis                  // do not automatically insert semicolons - for testing only
)

type Scanner struct {
	gombiScanner
	ErrorCount int // number of errors encountered

	lastIsPreSemi       bool
	commentAfterPreSemi bool
	endOfLinePos        int

	endOfLine int

	tokBuf tok

	file *token.File  // source file handle
	dir  string       // directory portion of file.Name()
	err  ErrorHandler // error reporting; or nil
	mode Mode         // scanning mode
	src  []byte
}

func skipBOM(buf []byte) []byte {
	r, size := utf8.DecodeRune(buf)
	if size > 0 && r == 0xFEFF {
		buf = buf[size:]
	}
	return buf
}

func (s *Scanner) Init(file *token.File, src []byte, err ErrorHandler, mode Mode) {
	//fmt.Println("Init src", strconv.Quote(string(src)), mode)

	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size(), len(src)))
	}

	s.gombiScanner = newGombiScanner()
	s.src = skipBOM(src)
	s.SetSource(s.src)

	s.file = file
	s.dir, _ = filepath.Split(file.Name())
	s.err = err
	s.mode = mode

	s.ErrorCount = 0

	s.lastIsPreSemi = false
	s.commentAfterPreSemi = false
	s.endOfLinePos = 0
	s.endOfLine = 0
}

var newlineValue = []byte{'\n'}

func (s *Scanner) insertSemi() (t tok, ok bool) {
	if s.mode&dontInsertSemis == 0 && s.lastIsPreSemi {
		t.id = int(token.SEMICOLON)
		t.val = newlineValue
		t.pos = s.endOfLinePos
		return t, true
	}
	return t, false
}

func (s *Scanner) addLineFromValue(pos int, val []byte) (added bool) {
	for i, c := range val {
		if c == '\n' {
			s.file.AddLine(pos + i + 1)
			added = true
		}
	}
	return
}

func isPreSemi(tok int) bool {
	switch token.Token(tok) {
	case token.IDENT, token.INT, token.FLOAT, token.IMAG, token.CHAR,
		token.STRING, token.BREAK, token.CONTINUE, token.FALLTHROUGH,
		token.RETURN, token.INC, token.DEC, token.RPAREN, token.RBRACK,
		token.RBRACE:
		return true
	}
	return false
}

func (s *Scanner) Scan() (token.Pos, token.Token, string) {
	for {
		var val []byte
		s.gombiScanner.Scan()
		t := s.Token()
		// supress value for operations
		if t.ID != int(token.SEMICOLON) && firstOp <= t.ID && t.ID <= lastOp {
			goto Return
		} else {
			st := s.Token()
			val = s.src[st.Lo:st.Hi]
		}
		switch t.ID {
		case tWhitespace: // skip whitespace
			if s.lastIsPreSemi && !s.commentAfterPreSemi {
				s.endOfLinePos = s.Pos() + 1
			}
			continue
		// add newline
		case tNewline:
			s.file.AddLine(t.Lo + 1)
			if s.mode&dontInsertSemis == 0 && s.lastIsPreSemi {
				t.ID = int(token.SEMICOLON)
				t.Lo = s.endOfLinePos
				val = newlineValue
				goto Return
			}
			continue
		case tLineComment:
			if s.mode&dontInsertSemis == 0 && s.lastIsPreSemi {
				s.SetPos(t.Lo)
				t.ID = int(token.SEMICOLON)
				t.Lo = s.endOfLinePos
				val = newlineValue
				goto Return
			}
			s.addLineFromValue(t.Lo, val)
			if s.lastIsPreSemi {
				s.commentAfterPreSemi = true
			}

			// modify
			t.ID = int(token.COMMENT)
			if val[len(val)-1] == '\n' {
				val = val[:len(val)-1]
			}
			val = stripCR(val)

			if s.mode&ScanComments == 0 {
				continue
			}
		case tGeneralCommentML:
			if s.mode&dontInsertSemis == 0 && s.lastIsPreSemi {
				s.SetPos(t.Lo)
				t.ID = int(token.SEMICOLON)
				t.Lo = s.endOfLinePos
				val = newlineValue
				goto Return
			}
			s.addLineFromValue(t.Lo, val)
			if s.lastIsPreSemi {
				s.commentAfterPreSemi = true
			}
			t.ID = int(token.COMMENT)
			val = stripCR(val)

			if s.mode&ScanComments == 0 {
				continue
			}
		case tGeneralCommentSL:
			if s.lastIsPreSemi {
				s.commentAfterPreSemi = true
			}
			t.ID = int(token.COMMENT)
			val = stripCR(val)
			oriPos := t.Lo
			if s.mode&dontInsertSemis == 0 && s.lastIsPreSemi {
				for {
					s.gombiScanner.Scan()
					t := s.Token()
					val = s.src[t.Lo:t.Hi]
					switch t.ID {
					case int(token.EOF), tNewline, tLineComment, tGeneralCommentML:
						s.SetPos(oriPos)
						t.ID = int(token.SEMICOLON)
						t.Lo = s.endOfLinePos
						val = newlineValue
						goto Return
					case tGeneralCommentSL:
					case tWhitespace:
					default:
						goto Return
					}
				}
			}
			if s.mode&ScanComments == 0 {
				continue
			}
		case tInterpretedStringLit:
			t.ID = int(token.STRING)
		case tRawStringLit:
			s.addLineFromValue(t.Lo, val)
			t.ID = int(token.STRING)
			val = stripCR(val)
		case int(token.EOF):
			if s.mode&dontInsertSemis == 0 && s.lastIsPreSemi {
				t.ID = int(token.SEMICOLON)
				t.Lo = s.endOfLinePos
				val = newlineValue
				goto Return
			}
		}
	Return:

		if t.ID != int(token.ILLEGAL) {
			s.endOfLinePos = s.Pos() + 1
			s.lastIsPreSemi = isPreSemi(t.ID)
			if s.lastIsPreSemi {
				s.commentAfterPreSemi = false
			}
		}
		return s.file.Pos(t.Lo), token.Token(t.ID), string(val)
	}
}

type Mode uint

type ErrorHandler func(pos token.Position, msg string)

func stripCR(b []byte) []byte {
	i := 0
	for _, ch := range b {
		if ch != '\r' {
			b[i] = ch
			i++
		}
	}
	return b[:i]
}

type tok struct {
	id  int
	pos int
	val []byte
}
