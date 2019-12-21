package main

import (
	"fmt"

	"github.com/growerlab/mensa/mensa"
)

func main() {
	fmt.Println("=================================")
	defer fmt.Println("=================================")
	fmt.Println(mensa.UA)
	fmt.Println("BuiltTime: ", mensa.BUILDTIME)
	fmt.Println("Commit: ", mensa.BUILDCOMMIT)
	fmt.Println("GoVersion: ", mensa.GOVERSION)

	if err := mensa.Run(); err != nil {
		panic(err)
	}
}
