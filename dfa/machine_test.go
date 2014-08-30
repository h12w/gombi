package dfa

import (
	"testing"

	"github.com/hailiang/gspec"
)

func TestMatch(t *testing.T) {
	expect := gspec.Expect(t.FailNow)
	m := threeToken()
	input := []byte("0x12A 123 abc")
	p := 0
	for _, testcase := range []struct {
		token string
		label int
	}{
		{"0x12A", hexLabel},
		{"123", decimalLabel},
		{"abc", identLabel},
	} {
		size, label, ok := m.Match(input[p:])
		expect("matched", ok).Equal(true)
		expect("matched label", label).Equal(testcase.label)
		expect("token", string(input[p:p+size])).Equal(testcase.token)
		p += size
		p += 1
	}
}
