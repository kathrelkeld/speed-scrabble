package main

import (
	"fmt"
)

func Example_PrintBoard() {
	b := make(Board, 2)
	b[0] = []string{"A", "B"}
	b[1] = []string{"", "C"}
	fmt.Println(b.String())
	// Output: AB
	//  C
}
