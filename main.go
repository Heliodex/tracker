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
Order: 0

Pattern 0
`[1:]

	tf, err := loadtext.ReadText(bytes.NewBuffer([]byte(testcontent)))
	if err != nil {
		panic(err)
	}

	fmt.Println(tf)
}
