package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/Heliodex/tracker/load"
)

func main() {
	content1, err := os.ReadFile("./track.xm")
	if err != nil {
		panic(err)
	}

	f, err := load.Read(bytes.NewBuffer(content1))
	if err != nil {
		panic(err)
	}

	fmt.Println(f)
}
