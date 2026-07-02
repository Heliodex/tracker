package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/Heliodex/tracker/load"
	"github.com/Heliodex/tracker/save"
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

	if _, err = r.Seek(0, 0); err != nil {
		panic(err)
	}

	content1, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	w := &bytes.Buffer{}
	if err = save.Write(w, f); err != nil {
		panic(err)
	}

	content2 := w.Bytes()

	for i := range content1 {
		// fmt.Printf("Byte %d: %d %d\n", i, content1[i], content2[i])
		if content1[i] != content2[i] {
			fmt.Printf("Byte %d differs: %d != %d\n", i, content1[i], content2[i])
			break
		}
	}

	fmt.Println("Lengths:", len(content1), len(content2))
}
