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
		switch t.ID {
		case tWhitespace:
			s.endOfLine = t.Hi + 1
			continue
		case tNewline:
			if s.preSemi && s.mode&dontInsertSemis == 0 {
				t.ID, t.Lo, val, s.preSemi = tSemi, s.endOfLine, newlineValue, false
				break
			}
			s.file.AddLine(t.Lo + 1)
			continue
		case tIdent, tInt, tFloat, tImag, tChar, tString, tReturn, tBreak, tContinue, tFallthrough:
			s.preSemi, s.endOfLine = true, t.Hi+1
			val = s.src[t.Lo:t.Hi]
		case tRParen, tRBrack, tRBrace, tInc, tDec:
			s.preSemi, s.endOfLine = true, t.Hi+1
		case tLineComment:
			if s.preSemi && s.mode&dontInsertSemis == 0 {
				s.SetPos(t.Lo)
				t.ID, t.Lo, val, s.preSemi = tSemi, s.endOfLine, newlineValue, false
				break
			}
			if s.mode&ScanComments == 0 {
				continue
			}
			t.ID = tComment
			val = s.src[t.Lo:t.Hi]
			if val[len(val)-1] == '\n' {
				s.file.AddLine(t.Hi)
				val = val[:len(val)-1]
			}
			val = stripCR(val)
		case tGeneralCommentML:
			if s.preSemi && s.mode&dontInsertSemis == 0 {
				s.SetPos(t.Lo)
				t.ID, t.Lo, val, s.preSemi = tSemi, s.endOfLine, newlineValue, false
				break
			}
			if s.mode&ScanComments == 0 {
				continue
			}
			t.ID = tComment
			val = s.src[t.Lo:t.Hi]
			for i, c := range val {
				if c == '\n' {
					s.file.AddLine(t.Lo + i + 1)
				}
			}
			val = stripCR(val)
		case tGeneralCommentSL:
			if s.preSemi && s.mode&dontInsertSemis == 0 {
				s.preSemi = false
				t = t.Copy()
				for s.gombiScanner.Scan() {
					nt := s.Token()
					switch nt.ID {
					case tEOF, tNewline, tLineComment, tGeneralCommentML:
						s.SetPos(t.Lo)
						t.ID, t.Lo, val = tSemi, s.endOfLine, newlineValue
						goto returnSemi
					case tWhitespace, tGeneralCommentSL:
						continue
					default:
						s.SetPos(t.Hi)
						goto returnComment
					}
				}
			returnSemi:
				break
			}
		returnComment:
			if s.mode&ScanComments == 0 {
				continue
			}
			t.ID = tComment
			val = stripCR(s.src[t.Lo:t.Hi])
		case tInterpretedStringLit:
			s.preSemi, s.endOfLine = true, t.Hi+1
			t.ID = tString
			val = s.src[t.Lo:t.Hi]
		case tRawStringLit:
			s.preSemi, s.endOfLine = true, t.Hi+1
			t.ID = tString
			val = s.src[t.Lo:t.Hi]
			for i, c := range val {
				if c == '\n' {
					s.file.AddLine(t.Lo + i + 1)
				}
			}
			val = stripCR(val)
		case tEOF:
			if s.preSemi && s.mode&dontInsertSemis == 0 {
				s.SetPos(t.Lo)
				t.ID, t.Lo, val, s.preSemi = tSemi, s.endOfLine, newlineValue, false
			}
		case tSemi:
			s.preSemi = false
			val = s.src[t.Lo:t.Hi]
		case tIllegal:
			val = s.src[t.Lo:t.Hi]
		default:
			s.preSemi = false
			if t.ID < firstOp || t.ID > lastOp {
				val = s.src[t.Lo:t.Hi]
			}
		}
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
