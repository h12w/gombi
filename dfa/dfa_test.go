package dfa

import (
	"fmt"

	"testing"
)

func Test(t *testing.T) {
	fmt.Println()
	ss := [6]state{}
	s0, s1, s2, s3, s4, s5 := &ss[0], &ss[1], &ss[2], &ss[3], &ss[4], &ss[5]
	s1.accept = true
	s2.accept = true
	s4.accept = true
	s5.accept = true
	s0.set('0', s1)
	s0.setBetween('1', '9', s2)
	s0.setBetween('A', 'Z', s5)
	s0.setBetween('a', 'z', s5)
	s1.set('x', s3)
	s1.setBetween('0', '9', s2)
	s2.setBetween('0', '9', s2)
	s3.setBetween('0', '9', s4)
	s3.setBetween('A', 'F', s4)
	s3.setBetween('a', 'f', s4)
	s4.setBetween('0', '9', s4)
	s4.setBetween('A', 'F', s4)
	s4.setBetween('a', 'f', s4)
	s5.setBetween('0', '9', s5)
	s5.setBetween('A', 'Z', s5)
	s5.setBetween('a', 'z', s5)
	for i := range ss {
		fmt.Print(ss[i].dump(s0))
	}
	fmt.Println((&dfa{ss[:]}).saveSvg("t.svg"))

	input := []byte("123")

	s := s0
	i := 0
	for i < len(input) {
		s = s.next(input[i])
		if s == nil {
			fmt.Println("illegal")
			break
		}
		i++
	}
	switch s {
	case s2:
		fmt.Println("digit")
	case s5:
		fmt.Println("ident")
	case s4:
		fmt.Println("hex")
	}
}
