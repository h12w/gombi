package scanner

import (
	"fmt"
	"go/token"
	"path/filepath"
	"unicode/utf8"

	"github.com/hailiang/gombi/scan"
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

	commentQueue tokenQueue
	endOfLine    int
	tokBuf       scan.Token

	file *token.File  // source file handle
	dir  string       // directory portion of file.Name()
	err  ErrorHandler // error reporting; or nil
	mode Mode         // scanning mode
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
	s.SetSource(skipBOM(src))

	s.file = file
	s.dir, _ = filepath.Split(file.Name())
	s.err = err
	s.mode = mode

	s.ErrorCount = 0

	s.lastIsPreSemi = false
	s.commentAfterPreSemi = false
	s.endOfLinePos = 0
	s.endOfLine = 0
	s.commentQueue.reset()
}

var newlineValue = []byte{'\n'}

func (s *Scanner) insertSemi() *scan.Token {
	if s.mode&dontInsertSemis == 0 &&
		s.lastIsPreSemi {
		s.tokBuf.ID = int(token.SEMICOLON)
		s.tokBuf.Value = newlineValue
		s.tokBuf.Pos = s.endOfLinePos
		return &s.tokBuf
	}
	return nil
}

var skipToken = &scan.Token{ID: tSkip}

func (s *Scanner) scanToken() *scan.Token {
	if s.commentQueue.count() > 0 {
		return s.commentQueue.pop()
	}
	s.gombiScanner.Scan()
	t := s.Token()
	// supress value for operations
	if t.ID != int(token.SEMICOLON) && firstOp <= t.ID && t.ID <= lastOp {
		t.Value = nil
		return t
	}
	switch token.Token(t.ID) {
	case tWhitespace: // skip whitespace
		if s.lastIsPreSemi && !s.commentAfterPreSemi {
			s.endOfLinePos = s.Pos() + 1
		}
		return skipToken
	// add newline
	case tNewline:
		s.file.AddLine(t.Pos + 1)
		if semi := s.insertSemi(); semi != nil {
			return semi
		}
		return skipToken
	case tLineComment:
		s.addLineFromValue()
		if s.lastIsPreSemi {
			s.commentAfterPreSemi = true
		}

		// modify
		t.ID = int(token.COMMENT)
		if t.Value[len(t.Value)-1] == '\n' {
			t.Value = t.Value[:len(t.Value)-1]
		}
		t.Value = stripCR(t.Value)

		if semi := s.insertSemi(); semi != nil {
			s.commentQueue.push(t)
			return semi
		}
		if s.mode&ScanComments == 0 {
			return skipToken
		}
	case tGeneralCommentML:
		s.addLineFromValue()
		if s.lastIsPreSemi {
			s.commentAfterPreSemi = true
		}
		t.ID = int(token.COMMENT)
		t.Value = stripCR(t.Value)
		if semi := s.insertSemi(); semi != nil {
			s.commentQueue.push(t)
			return semi
		}
		if s.mode&ScanComments == 0 {
			return skipToken
		}
	case tGeneralCommentSL:
		if s.lastIsPreSemi {
			s.commentAfterPreSemi = true
		}
		t.ID = int(token.COMMENT)
		t.Value = stripCR(t.Value)
		if semi := s.insertSemi(); semi != nil {
			for {
				s.commentQueue.push(t)
				s.gombiScanner.Scan()
				t = s.Token()
				switch t.ID {
				case int(token.EOF), tNewline, tLineComment, tGeneralCommentML:
					s.commentQueue.push(t)
					return semi
				case tGeneralCommentSL:
					s.commentQueue.push(t)
				case tWhitespace:
				default:
					s.commentQueue.push(t)
					return skipToken
				}
			}
		}
		if s.mode&ScanComments == 0 {
			return skipToken
		}
	case tInterpretedStringLit:
		t.ID = int(token.STRING)
	case tRawStringLit:
		s.addLineFromValue()
		t.ID = int(token.STRING)
		t.Value = stripCR(t.Value)
	case token.EOF:
		if semi := s.insertSemi(); semi != nil {
			return semi
		}
		return t
	}

	return t
}

func (s *Scanner) addLineFromValue() (added bool) {
	for i, c := range s.Token().Value {
		if c == '\n' {
			s.file.AddLine(s.Token().Pos + i + 1)
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
	if s.Token() != nil && s.Token().ID == int(token.EOF) {
		return s.file.Pos(s.Pos()), token.EOF, ""
	}
	var t *scan.Token
	for {
		t = s.scanToken()
		if t.ID != tSkip {
			break
		}
	}
	if t.ID != int(token.ILLEGAL) {
		s.endOfLinePos = s.Pos() + 1
		s.lastIsPreSemi = isPreSemi(t.ID)
		if s.lastIsPreSemi {
			s.commentAfterPreSemi = false
		}
	}
	return s.file.Pos(t.Pos), token.Token(t.ID), string(t.Value)
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

type tokenQueue struct {
	a []*scan.Token
}

func (q *tokenQueue) push(t *scan.Token) {
	q.a = append(q.a, t)
}

func (q *tokenQueue) pop() (t *scan.Token) {
	t, q.a = q.a[0], q.a[1:]
	return t
}

func (q *tokenQueue) count() int {
	return len(q.a)
}

func (q *tokenQueue) reset() {
	q.a = q.a[0:0]
}
