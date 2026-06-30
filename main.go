package main

import (
	"fmt"
	"os"

	"github.com/Heliodex/tracker/load"
)

func main() {
	r, err := os.Open("./track.xm")
	if err != nil {
		panic(err)
	}

	f, err := load.Read(r)
	if err != nil {
		panic(err)
	}

	fmt.Println(f)
}
