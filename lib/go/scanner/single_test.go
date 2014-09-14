package scanner

import (
	"fmt"
	"go/token"
	"io/ioutil"
	"runtime"
	"sort"
	"testing"

	"github.com/hailiang/gombi/scan"
)

var sampleGoFile = runtime.GOROOT() + "/src/pkg/go/scanner/scanner.go"

func TestSingle(t *testing.T) {
	//	fmt.Println(int(token.INT))
}

type sortItem struct {
	count int
	tok   token.Token
}
type sortItems []sortItem

func (rs sortItems) Len() int           { return len(rs) }
func (rs sortItems) Less(i, j int) bool { return rs[i].count < rs[j].count }
func (rs sortItems) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

func ATestCount(t *testing.T) {
	src, err := ioutil.ReadFile(sampleGoFile)
	if err != nil {
		panic(err)
	}
	fset := token.NewFileSet()
	file := fset.AddFile(sampleGoFile, fset.Base(), len(src))
	var s Scanner
	s.Init(file, src, nil, ScanComments)
	m := make(map[token.Token]int)
	for {
		_, tok, _ := s.Scan()
		m[tok]++
		if tok == token.EOF {
			break
		}
	}
	items := sortItems{}
	for k, c := range m {
		items = append(items, sortItem{c, k})
	}
	sort.Sort(items)
	for _, item := range items {
		fmt.Println(item.count, item.tok)
	}
}

func BenchmarkSpec(b *testing.B) {
	for i := 0; i < b.N; i++ {
		spec()
	}
}

func test() {
	var (
		c     = scan.Char
		or    = scan.Or
		class = scan.CharClass

		unicodeLetter = class(`L`)
		unicodeDigit  = class(`Nd`)
		letter        = or(unicodeLetter, c(`_`))

		ident = or(letter, unicodeDigit)
	)
	_ = ident
}
