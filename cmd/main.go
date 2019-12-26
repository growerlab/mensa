package main

import (
	"fmt"

	"github.com/growerlab/mensa/mensa"
)

func main() {
	fmt.Println("=================================")
	fmt.Println(mensa.UA)
	fmt.Println("BuiltTime: ", mensa.BUILDTIME)
	fmt.Println("CommitID: ", mensa.BUILDCOMMIT)
	fmt.Println("GoVersion: ", mensa.GOVERSION)
	fmt.Println("=================================")

	if err := mensa.Run(); err != nil {
		panic(err)
	}
}
