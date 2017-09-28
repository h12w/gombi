package scanner

import (
	"go/token"
	"io/ioutil"
	"runtime"
	"sort"
	"testing"
)

var sampleGoFile = runtime.GOROOT() + "/src/go/scanner/scanner.go"

func TestSingle(t *testing.T) {
	//s := "'\\000'"
	//fmt.Println(runeLit.Match([]byte(s)))
	//	fmt.Println(s(`'`).Match([]byte("'")))
}

type sortItem struct {
	count int
	tok   token.Token
}
type sortItems []sortItem

func (rs sortItems) Len() int           { return len(rs) }
func (rs sortItems) Less(i, j int) bool { return rs[i].count < rs[j].count }
func (rs sortItems) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

func TestCount(t *testing.T) {
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
	//for _, item := range items {
	//	fmt.Println(item.count, item.tok)
	//}
}
