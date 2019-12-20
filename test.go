package main

import (
	"fmt"
	"net/url"
	"strings"
)

func main() {
	a, _ := url.Parse("https://github.com/growerlab/mensa.git")
	fmt.Println(a.Path)

	path := a.Path

	aa := strings.FieldsFunc(path, func(r rune) bool {
		return r == rune('/') || r == rune('.')
	})

	fmt.Println(aa)
}
