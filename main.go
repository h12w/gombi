package main

import (
	"fmt"
	"unicode"
)

type (
	ParserRune   func([]rune) (rune, []rune, error)
	ParserString func([]rune) (string, []rune, error)
)

func ConRune(ps ...ParserRune) ParserString {
	return func(input []rune) (string, []rune, error) {
		result := []rune{}
		remain := input
		for _, p := range ps {
			var r rune
			var err error
			if r, remain, err = p(input); err != nil {
				return "", input, err
			}
			result = append(result, r)
		}
		return string(result), remain, nil
	}
}

func Letter(input []rune) (rune, []rune, error) {
	if len(input) > 0 && unicode.IsLetter(input[0]) {
		return input[0], input[1:], nil
	}
	return 0, input, fmt.Errorf("expect letter")
}

func main() {
	p("Hello World!")
}

func p(v ...interface{}) {
	fmt.Println(v...)
}
