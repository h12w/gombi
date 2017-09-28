package ua

import (
	"io"

	"h12.me/gombi/experiment/gre/scan"
	"h12.me/gombi/parse"
)

var (
	userAgent      = rule("user-agent", or(product, comment).As("item").OneOrMore())
	product        = rule("product", productName, con(productSep, productVersion).ZeroOrOne())
	productName    = rule("product-name", productToken)
	productVersion = rule("product-version", productToken)
	productToken   = term("product-token")
	productSep     = term("/")
	comment        = rule("comment", leftParen, or(commentText, self("comment")).ZeroOrMore(), rightParen)
	leftParen      = term(`"("`)
	rightParen     = term(`")"`)
	commentText    = term("ctext")
	parser         = parse.NewParser(userAgent)
	scanner        = newScanner()
)

const (
	tEOF = iota
	tProductToken
	tProductSep
	tLWS
	tLeftParen
	tRightParen
	tCommentSep
	tCommentText
	tokenCount
)

var (
	TokenTable = TTable(tokenCount, []TT{
		{tProductToken, productToken},
		{tProductSep, productSep},
		{tLeftParen, leftParen},
		{tRightParen, rightParen},
		{tCommentText, commentText},
		{tEOF, parse.EOF},
	})
)

func newScanner() *switchScanner {
	var (
		c     = scan.Char
		merge = scan.Merge
		or    = scan.Or
		con   = scan.Con

		CHAR  = c(`\x00-\x7F`)
		OCTET = c(`\x00-\xFF`)
		CR    = c(`\r`)
		LF    = c(`\n`)
		CRLF  = con(CR, LF)
		SP    = c(` `)
		HT    = c(`\t`)
		LWS   = con(CRLF.ZeroOrOne(), or(SP, HT).OneOrMore())
		CTL   = c(`\x00-\x1F\x7F`)
		//TEXT  = or(OCTET.Exclude(CTL), LWS)

		separators   = merge(c(`\(\)<>@,;:\\"/\[\]\?=\{\}`), SP, HT)
		token        = CHAR.Exclude(CTL, separators).OneOrMore()
		quotedPair   = con(c(`\\`), CHAR)
		ctext        = or(OCTET.Exclude(CTL, c(`()`), c(`;`)), LWS)
		leftParen    = c(`(`)
		rightParen   = c(`)`)
		productToken = token
		productSep   = c(`/`)
		commentSep   = or(c(`;`), LWS).OneOrMore()
		commentText  = or(ctext, quotedPair).OneOrMore()

		m = scan.NewMatcher(
			productToken,
			productSep,
			LWS,
			leftParen,
		)
		mc = scan.NewMapMatcher(scan.MM{
			{leftParen, tLeftParen},
			{rightParen, tRightParen},
			{commentSep, tCommentSep},
			{commentText, tCommentText},
		})

		scanner = &switchScanner{scan.Scanner{Matcher: m}, m, mc, 0}
	)
	return scanner
}

type switchScanner struct {
	scan.Scanner
	m      *scan.Matcher
	mc     *scan.Matcher
	clevel int
}

func (s *switchScanner) parserToken() (*parse.Token, *parse.R) {
	t := s.Scanner.Token()
	return &parse.Token{t.Value, t.Pos}, TokenTable[t.ID]
}

func (s *switchScanner) SetReader(r io.Reader) {
	s.Matcher = s.m
	s.Scanner.SetReader(r)
}

func (s *switchScanner) Scan() bool {
	if s.Scanner.Scan() {
		switch s.Token().ID {
		case tLeftParen:
			s.clevel++
			if s.clevel > 0 {
				s.Matcher = s.mc
			}
		case tRightParen:
			s.clevel--
			if s.clevel == 0 {
				s.Matcher = s.m
			}
		case tLWS, tCommentSep:
			return s.Scan() // skip
		}
		return true
	}
	return false
}

type TT struct {
	Token int
	Term  *parse.R
}

func TTable(size int, tt []TT) []*parse.R {
	t := make([]*parse.R, size)
	for i := range tt {
		t[tt[i].Token] = tt[i].Term
	}
	return t
}
