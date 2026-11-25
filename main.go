package main

import (
	"bytes"
	"fmt"

	"github.com/Heliodex/tracker/loadtext"
)

func main() {
	// content1, err := os.ReadFile("./track2.xm")
	// if err != nil {
	// 	panic(err)
	// }

	// f, err := load.Read(bytes.NewBuffer(content1))
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(f)
	testcontent := `
Name: cool guy
NumChannels: 1
NumPatterns: 1
Order: 0

Pattern 0
Rows: 1
Row 000: C-4 01 .. .. ..
`[1:]

	tf, err := loadtext.ReadText(bytes.NewBuffer([]byte(testcontent)))
	if err != nil {
		panic(err)
	}

	fmt.Println(tf)
}
