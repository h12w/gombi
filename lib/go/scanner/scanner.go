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

func (s *Scanner) insertSemi() *scan.Token {
	if s.mode&dontInsertSemis == 0 &&
		s.lastIsPreSemi {
		return &scan.Token{
			ID:    int(token.SEMICOLON),
			Value: []byte{'\n'},
			Pos:   s.endOfLinePos,
		}
	}
	return nil
}

var skipToken = &scan.Token{ID: tSkip}

func (s *Scanner) scan() *scan.Token {
	if s.commentQueue.count() > 0 {
		return s.commentQueue.pop()
	}
	s.gombiScanner.Scan()
	t := s.Token()
	//fmt.Println("scanning:", t.ID, strconv.Quote(string(t.Value)), s.lastIsPreSemi)

	// add line
	switch token.Token(t.ID) {
	case tNewline:
		s.file.AddLine(t.Pos + 1)
	case tLineComment, tGeneralCommentML, tRawStringLit:
		s.addLineFromValue(t.Pos, t.Value)
	}

	if s.lastIsPreSemi {
		switch token.Token(t.ID) {
		case tLineComment, tGeneralCommentSL, tGeneralCommentML:
			s.commentAfterPreSemi = true
		}
	}

	// skip whitespace
	if t.ID == int(tWhitespace) {
		if s.lastIsPreSemi && !s.commentAfterPreSemi {
			s.endOfLinePos = s.Pos() + 1
		}
		return skipToken
	}

	// supress value for operations
	if t.ID != int(token.SEMICOLON) && firstOp <= t.ID && t.ID <= lastOp {
		t.Value = nil
		return t
	}

	// insert semi

	switch token.Token(t.ID) {
	case tNewline:
		if semi := s.insertSemi(); semi != nil {
			return semi
		}
		return skipToken
	case token.EOF:
		if semi := s.insertSemi(); semi != nil {
			return semi
		}
		return t
	case tLineComment, tGeneralCommentML:
		if semi := s.insertSemi(); semi != nil {
			modify(t)
			s.commentQueue.push(t)
			return semi
		}
	case tGeneralCommentSL:
		if semi := s.insertSemi(); semi != nil {
			for {
				modify(t)
				s.commentQueue.push(t)
				s.gombiScanner.Scan()
				t = s.Token()
				//fmt.Println("SL scanning:", t.ID, strconv.Quote(string(t.Value)), s.lastIsPreSemi)
				switch t.ID {
				case int(token.EOF), tNewline, tLineComment, tGeneralCommentML:
					modify(t)
					s.commentQueue.push(t)
					return semi
				case tGeneralCommentSL:
					modify(t)
					s.commentQueue.push(t)
				case tWhitespace:
				default:
					modify(t)
					s.commentQueue.push(t)
					return skipToken
				}
			}
		}
	}

	// skip comments
	if s.mode&ScanComments == 0 {
		switch token.Token(t.ID) {
		case tLineComment, tGeneralCommentSL, tGeneralCommentML:
			return skipToken
		}
	}

	modify(t)
	return t
}

func modify(t *scan.Token) {
	// modify tokens
	switch token.Token(t.ID) {
	case tLineComment:
		t.ID = int(token.COMMENT)
		if t.Value[len(t.Value)-1] == '\n' {
			t.Value = t.Value[:len(t.Value)-1]
		}
		t.Value = stripCR(t.Value)
	case tGeneralCommentSL:
		t.ID = int(token.COMMENT)
		t.Value = stripCR(t.Value)
	case tGeneralCommentML:
		t.ID = int(token.COMMENT)
		t.Value = stripCR(t.Value)
	case tRawStringLit:
		t.ID = int(token.STRING)
		t.Value = stripCR(t.Value)
	case tInterpretedStringLit:
		t.ID = int(token.STRING)
	}

}

func (s *Scanner) addLineFromValue(pos int, val []byte) (added bool) {
	for i, c := range val {
		if c == '\n' {
			s.file.AddLine(pos + i)
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
	defer func() {
		if t.ID != int(token.ILLEGAL) {
			s.endOfLinePos = s.Pos() + 1
			s.lastIsPreSemi = isPreSemi(t.ID)
		}
		if s.lastIsPreSemi {
			s.commentAfterPreSemi = false
		}
		//fmt.Println("scan return:", t.ID, strconv.Quote(string(t.Value)), s.lastIsPreSemi)
	}()

	for {
		t = s.scan()
		if t.ID == int(token.ILLEGAL) {
			//fmt.Println("scan error", s.Error()) // DEBUG
			break
		} else if t.ID == tSkip {
			continue
		}
		break
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
