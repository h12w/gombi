package dfa

/*
func Test(t *testing.T) {
	m := threeToken()
	ss := m.ss
	s := &m.ss[0]
	i := 0
	input := []byte("0x12A")
	for i < len(input) {
		s = &m.ss[s.next(input[i])]
		if s == nil {
			fmt.Println("illegal")
			break
		}
		i++
	}
	switch s {
	case &ss[2]:
		fmt.Println("digit")
	case &ss[5]:
		fmt.Println("ident")
	case &ss[4]:
		fmt.Println("hex")
	}
}
*/
