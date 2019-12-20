package main

import "github.com/growerlab/mensa/mensa"

func main() {
	//fmt.Fprintf(os.Stderr, "%v", me)
	if err := mensa.Run(); err != nil {
		panic(err)
	}
} 
