package scanner

import (
	"fmt"
	"path/filepath"
	"unicode/utf8"

	"go/token"

	"github.com/hailiang/gombi/scan"
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
				t.ID, t.Lo, val, s.preSemi = tSemiColon, s.endOfLine, newlineValue, false
				break
			}
			s.file.AddLine(t.Lo + 1)
			continue
		case tIdentifier, tInt, tFloat, tImag, tRune, tString, tReturn, tBreak, tContinue, tFallthrough:
			s.preSemi, s.endOfLine = true, t.Hi+1
			val = s.src[t.Lo:t.Hi]
		case tRightParen, tRightBrack, tRightBrace, tInc, tDec:
			s.preSemi, s.endOfLine = true, t.Hi+1
		case tLineComment:
			if s.preSemi && s.mode&dontInsertSemis == 0 {
				s.SetPos(t.Lo)
				t.ID, t.Lo, val, s.preSemi = tSemiColon, s.endOfLine, newlineValue, false
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
				t.ID, t.Lo, val, s.preSemi = tSemiColon, s.endOfLine, newlineValue, false
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
					case tWhitespace, tGeneralCommentSL:
						continue
					case tEOF, tNewline, tLineComment, tGeneralCommentML:
						s.SetPos(t.Lo)
						t.ID, t.Lo, val = tSemiColon, s.endOfLine, newlineValue
						goto returnSemi
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
				t.ID, t.Lo, val, s.preSemi = tSemiColon, s.endOfLine, newlineValue, false
			}
		case tSemiColon:
			s.preSemi = false
			val = s.src[t.Lo:t.Hi]
		case eRune:
			val = s.src[t.Lo:t.Hi]
			t.ID = tRune
			s.error(t, "illegal rune literal")
		case eRuneEscapeChar:
			r := decodeRune(s.src[t.Hi-2:])
			val = s.src[t.Lo:t.Hi]
			t.ID = tRune
			t.Lo = t.Hi - 1
			s.error(t, fmt.Sprintf("illegal character %#U in escape sequence", r))
		case eRuneUnknownEscape:
			val = s.src[t.Lo:t.Hi]
			t.ID = tRune
			t.Lo += 2
			s.error(t, "unknown escape sequence")
		case eIllegal:
			r := decodeRune(s.src[t.Lo:])
			val = s.src[t.Lo : t.Lo+1]
			s.error(t, fmt.Sprintf("illegal character %#U", r))
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

func (s *Scanner) error(t *scan.Token, msg string) {
	if s.err != nil {
		s.err(s.file.Position(token.Pos(s.fileBase+t.Lo)), msg)
	}
}

func decodeRune(bs []byte) rune {
	r, _ := utf8.DecodeRune(bs)
	if r == utf8.RuneError {
		return rune(bs[0])
	}
	return r
}
