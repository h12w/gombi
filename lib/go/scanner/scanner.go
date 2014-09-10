package scanner

import (
	"fmt"
	"path/filepath"
	"unicode/utf8"

	"go/token"
)

const (
	ScanComments    Mode = 1 << iota // return comments as COMMENT tokens
	dontInsertSemis                  // do not automatically insert semicolons - for testing only
)

var newlineValue = []byte{'\n'}

type Scanner struct {
	gombiScanner
	ErrorCount int // number of errors encountered

	preSemi   bool
	endOfLine int

	file     *token.File // source file handle
	fileBase int
	dir      string       // directory portion of file.Name()
	err      ErrorHandler // error reporting; or nil
	mode     Mode         // scanning mode
	src      []byte
}
type Mode uint
type ErrorHandler func(pos token.Position, msg string)

func (s *Scanner) Init(file *token.File, src []byte, err ErrorHandler, mode Mode) {
	//fmt.Println("Init src", strconv.Quote(string(src)), mode)

	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size(), len(src)))
	}

	s.gombiScanner = newGombiScanner()
	s.src = skipBOM(src)
	s.SetSource(s.src)

	s.file = file
	s.fileBase = s.file.Base()
	s.dir, _ = filepath.Split(file.Name())
	s.err = err
	s.mode = mode

	s.ErrorCount = 0

	s.preSemi = false
	s.endOfLine = 0
}
func skipBOM(buf []byte) []byte {
	r, size := utf8.DecodeRune(buf)
	if size > 0 && r == 0xFEFF {
		buf = buf[size:]
	}
	return buf
}

func (s *Scanner) Scan() (token.Pos, token.Token, string) {
	for s.gombiScanner.Scan() {
		var val []byte
		t := s.Token()
		if firstOp <= t.ID && t.ID <= lastOp && t.ID != tSemi {
			goto PreReturn
		}
		switch t.ID {
		case tWhitespace:
			continue
		case tNewline:
			if s.preSemi && s.mode&dontInsertSemis == 0 {
				t.ID, t.Lo, val = tSemi, s.endOfLine, newlineValue
				goto PreReturn
			}
			s.file.AddLine(t.Lo + 1)
			continue
		case tLineComment:
			if s.preSemi && s.mode&dontInsertSemis == 0 {
				s.SetPos(t.Lo)
				t.ID, t.Lo, val = tSemi, s.endOfLine, newlineValue
				goto PreReturn
			}
			val = s.src[t.Lo:t.Hi]
			t.ID = tComment
			if val[len(val)-1] == '\n' {
				s.file.AddLine(t.Hi)
				val = val[:len(val)-1]
			}
			val = stripCR(val)

			if s.mode&ScanComments == 0 {
				continue
			}
		case tGeneralCommentML:
			if s.preSemi && s.mode&dontInsertSemis == 0 {
				s.SetPos(t.Lo)
				t.ID, t.Lo, val = tSemi, s.endOfLine, newlineValue
				goto PreReturn
			}
			val = s.src[t.Lo:t.Hi]
			for i, c := range val {
				if c == '\n' {
					s.file.AddLine(t.Lo + i + 1)
				}
			}
			t.ID = tComment
			val = stripCR(val)

			if s.mode&ScanComments == 0 {
				continue
			}
		case tGeneralCommentSL:
			t.ID = tComment
			oriPos := t.Lo
			if s.preSemi && s.mode&dontInsertSemis == 0 {
				for s.gombiScanner.Scan() {
					t := s.Token()
					switch t.ID {
					case tEOF, tNewline, tLineComment, tGeneralCommentML:
						s.SetPos(oriPos)
						t.ID, t.Lo, val = tSemi, s.endOfLine, newlineValue
						goto PreReturn
					case tWhitespace, tGeneralCommentSL:
					default:
						val = s.src[t.Lo:t.Hi]
						goto PreReturn
					}
				}
			}
			val = stripCR(s.src[t.Lo:t.Hi])
			if s.mode&ScanComments == 0 {
				continue
			}
		case tInterpretedStringLit:
			t.ID = tString
			val = s.src[t.Lo:t.Hi]
		case tRawStringLit:
			val = s.src[t.Lo:t.Hi]
			for i, c := range val {
				if c == '\n' {
					s.file.AddLine(t.Lo + i + 1)
				}
			}
			t.ID = tString
			val = stripCR(val)
		case tEOF:
			if s.preSemi && s.mode&dontInsertSemis == 0 {
				t.ID, t.Lo, val = tSemi, s.endOfLine, newlineValue
				goto PreReturn
			}
		case tIllegal:
			goto DirectReturn
		default:
			val = s.src[t.Lo:t.Hi]
		}

	PreReturn:
		s.endOfLine = s.Pos() + 1
		switch token.Token(t.ID) {
		case token.IDENT, token.RPAREN, token.RBRACK, token.RBRACE,
			token.INT, token.FLOAT, token.IMAG, token.CHAR, token.STRING,
			token.RETURN, token.INC, token.DEC,
			token.BREAK, token.CONTINUE, token.FALLTHROUGH:
			s.preSemi = true
		default:
			s.preSemi = false

		}
	DirectReturn:
		return token.Pos(s.fileBase + t.Lo), token.Token(t.ID), string(val)
	}
	return token.Pos(s.fileBase + len(s.src)), token.EOF, ""
}
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
