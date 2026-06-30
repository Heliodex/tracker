package main

import (
	"fmt"

	"github.com/gotracker/playback/format/xm"
)

func main() {
	data, err := xm.XM.Load("./track.xm", nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(data)
}
