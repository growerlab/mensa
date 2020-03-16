package main

import (
	"fmt"
	"github.com/growerlab/mensa/src"
)

func main() {
	fmt.Println("=================================")
	fmt.Println("BuiltTime: ", src.BUILDTIME)
	fmt.Println("CommitID: ", src.BUILDCOMMIT)
	fmt.Println("GoVersion: ", src.GOVERSION)
	fmt.Println("=================================")
	fmt.Println(src.UA)

	src.Run()
}
