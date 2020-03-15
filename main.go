package main

import (
	"fmt"

	"github.com/growerlab/mensa/mensa"
)

func main() {
	fmt.Println("=================================")
	fmt.Println("BuiltTime: ", mensa.BUILDTIME)
	fmt.Println("CommitID: ", mensa.BUILDCOMMIT)
	fmt.Println("GoVersion: ", mensa.GOVERSION)
	fmt.Println("=================================")
	fmt.Println(mensa.UA)

	mensa.Run()
}
