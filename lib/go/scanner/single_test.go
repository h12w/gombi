package scanner

import (
	"fmt"
	"runtime"
	"testing"
)

var sampleGoFile = runtime.GOROOT() + "/src/pkg/go/scanner/scanner.go"

func TestSingle(t *testing.T) {
	fmt.Println(getMatcher().Size())
	//s := "'\\000'"
	//fmt.Println(runeLit.Match([]byte(s)))
	//	fmt.Println(s(`'`).Match([]byte("'")))
}
