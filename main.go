package main

import (
	"fmt"
	"os"
)

func main() {
	r, err := os.Open("./track.xm")
	if err != nil {
		panic(err)
	}

	f, err := Read(r)
	if err != nil {
		panic(err)
	}

	fmt.Println(f)
}
