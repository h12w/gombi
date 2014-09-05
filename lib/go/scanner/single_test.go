package scanner

import (
	"runtime"
	"testing"
)

var sampleGoFile = runtime.GOROOT() + "/src/pkg/go/scanner/scanner.go"

func TestGen(t *testing.T) {
	//scan.GenGo(tokenDefs(), "dfa.go", "scanner")
}

func TestSingle(t *testing.T) {
	//	fmt.Println(getMatcher().Size())
}
