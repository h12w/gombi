package scanner

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
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
	tokScanner scan.Scanner
	mode       Mode // scanning mode
	src        []byte
	preSemi    bool
	semiPos    int

	file      *token.File // source file handle
	fileBase  int         // cache of file.Base()
	dir       string      // cache of the directory portion of file.Name()
	lineStart int         // record the start position of each line

	errScanner scan.Scanner
	ErrorCount int          // number of errors encountered
	err        ErrorHandler // error reporting; or nil
}
type Mode uint
type ErrorHandler func(pos token.Position, msg string)

func (s *Scanner) Init(file *token.File, src []byte, err ErrorHandler, mode Mode) {
	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size(), len(src)))
	}

	s.tokScanner = scan.Scanner{Matcher: getTokenMatcher()}
	s.errScanner = scan.Scanner{Matcher: getErrorMatcher()}
	s.src = skipBOM(src)
	s.tokScanner.SetSource(s.src)
	s.errScanner.SetSource(s.src)

	s.file = file
	s.fileBase = s.file.Base()
	s.dir, _ = filepath.Split(file.Name())
	s.err = err
	s.mode = mode

	s.ErrorCount = 0

	s.preSemi = false
	s.semiPos = 0
}
func skipBOM(buf []byte) []byte {
	r, size := utf8.DecodeRune(buf)
	if size > 0 && r == 0xFEFF {
		buf = buf[size:]
	}
	return buf
}

func (s *Scanner) Scan() (token.Pos, token.Token, string) {
	for s.tokScanner.Scan() {
		var val []byte
		t := s.tokScanner.Token()
		//fmt.Println(token.Token(t.ID), t, string(s.src[t.Lo:t.Hi]))
		switch t.ID {
		case tWhitespace:
			s.semiPos = t.Hi + 1
			continue
		case tNewline:
			if s.preSemi {
				s.preSemi = false
				if s.mode&dontInsertSemis == 0 {
					t.ID, t.Lo, val = tSemiColon, s.semiPos, newlineValue
					break
				}
			}
			s.file.AddLine(t.Hi)
			s.lineStart = t.Hi
			continue
		case tIdentifier, tInt, tFloat, tImag, tRune, tString, tReturn, tBreak, tContinue, tFallthrough:
			s.preSemi, s.semiPos = true, t.Hi+1
			val = s.src[t.Lo:t.Hi]
		case tRightParen, tRightBrack, tRightBrace, tInc, tDec:
			s.preSemi, s.semiPos = true, t.Hi+1
		case tLineComment, tLineCommentEOF, tLineCommentInfo:
			if s.preSemi {
				s.preSemi = false
				if s.mode&dontInsertSemis == 0 {
					s.tokScanner.SetPos(t.Lo)
					t.ID, t.Lo, val = tSemiColon, s.semiPos, newlineValue
					break
				}
			}
			val = s.src[t.Lo:t.Hi]
			if t.ID == tLineCommentInfo && t.Lo == s.lineStart {
				s.interpretLineComment(val, t.Hi)
			}
			if val[len(val)-1] == '\n' {
				s.file.AddLine(t.Hi)
				s.lineStart = t.Hi
				val = val[:len(val)-1]
			}
			if s.mode&ScanComments == 0 {
				continue
			}
			if t.ID == tLineCommentEOF && t.Hi < len(s.src) {
				t, val = s.handleError(t.Lo, t.Hi)
				break
			}
			t.ID = tComment
			val = stripCR(val)
		case tGeneralCommentML:
			if s.preSemi {
				s.preSemi = false
				if s.mode&dontInsertSemis == 0 {
					s.tokScanner.SetPos(t.Lo)
					t.ID, t.Lo, val = tSemiColon, s.semiPos, newlineValue
					break
				}
			}
			val = s.src[t.Lo:t.Hi]
			for i, c := range val {
				if c == '\n' {
					s.file.AddLine(t.Lo + i + 1)
				}
			}
			if s.mode&ScanComments == 0 {
				continue
			}
			t.ID = tComment
			val = stripCR(val)
		case tGeneralCommentSL:
			if s.preSemi {
				s.preSemi = false
				if s.mode&dontInsertSemis == 0 {
					t = t.Copy()
					for s.tokScanner.Scan() {
						nt := s.tokScanner.Token()
						switch nt.ID {
						case tWhitespace, tGeneralCommentSL:
							continue
						case tEOF, tNewline, tLineComment, tLineCommentEOF,
							tLineCommentInfo, tGeneralCommentML, eCommentIncomplete:
							s.tokScanner.SetPos(t.Lo)
							t.ID, t.Lo, val = tSemiColon, s.semiPos, newlineValue
							goto returnSemi
						default:
							s.tokScanner.SetPos(t.Hi)
							goto returnComment
						}
					}
				returnSemi:
					break
				}
			}
		returnComment:
			if s.mode&ScanComments == 0 {
				continue
			}
			t.ID = tComment
			val = stripCR(s.src[t.Lo:t.Hi])
		case tInterpretedStringLit:
			s.preSemi, s.semiPos = true, t.Hi+1
			t.ID = tString
			val = s.src[t.Lo:t.Hi]
		case tRawStringLit:
			s.preSemi, s.semiPos = true, t.Hi+1
			t.ID = tString
			val = s.src[t.Lo:t.Hi]
			for i, c := range val {
				if c == '\n' {
					s.file.AddLine(t.Lo + i + 1)
				}
			}
			val = stripCR(val)
		case tEOF:
			if s.preSemi {
				s.preSemi = false
				if s.mode&dontInsertSemis == 0 {
					s.tokScanner.SetPos(t.Lo)
					t.ID, t.Lo, val = tSemiColon, s.semiPos, newlineValue
				}
			}
		case tSemiColon:
			s.preSemi = false
			val = s.src[t.Lo:t.Hi]
		case eCommentIncomplete:
			if s.preSemi {
				s.preSemi = false
				if s.mode&dontInsertSemis == 0 {
					s.tokScanner.SetPos(t.Lo)
					t.ID, t.Lo, val = tSemiColon, s.semiPos, newlineValue
					break
				}
			}
			t.ID = tComment
			val = s.src[t.Lo:t.Hi]
			s.error(t.Lo, "comment not terminated")
		case eOctalLit:
			t.ID = tInt
			val = s.src[t.Lo:t.Hi]
			s.error(t.Lo, "illegal octal number")
		case eHexLit:
			t.ID = tInt
			val = s.src[t.Lo:t.Hi]
			s.error(t.Lo, "illegal hexadecimal number")
		case eIllegal:
			t, val = s.handleError(t.Lo, t.Hi)
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
func (s *Scanner) interpretLineComment(text []byte, pos int) {
	// get filename and line number, if any
	if i := bytes.LastIndex(text, []byte{':'}); i > 0 {
		if line, err := strconv.Atoi(strings.TrimSpace(string(text[i+1:]))); err == nil && line > 0 {
			// valid //line filename:line comment
			filename := string(bytes.TrimSpace(text[len(commentInfoPrefix):i]))
			if filename != "" {
				filename = filepath.Clean(filename)
				if !filepath.IsAbs(filename) {
					// make filename relative to current directory
					filename = filepath.Join(s.dir, filename)
				}
			}
			// update scanner position
			s.file.AddLineInfo(pos, filename, line) // +len(text)+1 since comment applies to next line
		}
	}
}

var commentInfoPrefix = []byte("//line ")

func (s *Scanner) handleError(pos, errPos int) (t *scan.Token, val []byte) {
	t = s.errScanner.Token()
	s.errScanner.SetPos(pos)
	s.errScanner.Scan()
	//fmt.Println(t.ID, string(s.src[t.Lo:t.Hi]))
	switch t.ID {
	case eEscape:
		r := decodeRune(s.src[errPos:])
		t.ID = tRune
		s.error(errPos, fmt.Sprintf("illegal character %#U in escape sequence", r))
	case eEscapeUnknown:
		t.ID = tRune
		s.error(errPos, "unknown escape sequence")
	case eIncompleteEscape:
		t.ID = tRune
		s.error(errPos, "escape sequence not terminated")
	case eRuneIncomplete:
		t.ID = tRune
		s.error(pos, "rune literal not terminated")
	case eRuneBOM:
		t.ID = tRune
		s.error(pos+1, "illegal byte order mark")
	case eRune:
		t.ID = tRune
		s.error(pos, "illegal rune literal")
	case eEscapeBigU:
		t.ID = tRune
		s.error(errPos-1, "escape sequence is invalid Unicode code point")
	case eStrIncomplete:
		t.ID = tString
		s.error(pos, "string literal not terminated")
	case eRawStrIncomplete:
		t.ID = tString
		s.error(pos, "raw string literal not terminated")
	case eStrWithNUL:
		t.ID = tString
		s.error(errPos, "illegal character NUL")
	case eStrWithBOM:
		t.ID = tString
		s.error(errPos-2, "illegal byte order mark")
	case eStrWithWrongUTF8:
		t.ID = tString
		s.error(errPos, "illegal UTF-8 encoding")
	case eCommentBOM:
		t.ID = tComment
		s.error(errPos, "illegal byte order mark")
	default:
		t.ID = eIllegal
		r := decodeRune(s.src[pos:])
		switch r {
		case 0xFEFF:
			t.Hi += len([]byte("\uFEFF"))
			s.error(t.Hi, "illegal byte order mark")
		default:
			t.Hi = pos + 1
			s.error(pos, fmt.Sprintf("illegal character %#U", r))
		}
	}
	val = s.src[t.Lo:t.Hi]
	s.tokScanner.SetPos(t.Hi)
	return
}

func (s *Scanner) error(errPos int, msg string) {
	s.ErrorCount++
	if s.err != nil {
		s.err(s.file.Position(token.Pos(s.fileBase+errPos)), msg)
	}
}

func decodeRune(bs []byte) rune {
	r, _ := utf8.DecodeRune(bs)
	if r == utf8.RuneError {
		return rune(bs[0])
	}
	return r
}
