package main

import (
	"fmt"
)

func ExampleHashPrefixToNumber() {
	var h Hash
	fmt.Println(h.PrefixToNumber(0))
	fmt.Println(h.PrefixToNumber(2))
	h[0] = 0b00100100
	h[1] = 0b00111111
	h[2] = 0b01101010
	h[3] = 0b10001000
	f := "%b\n"
	fmt.Printf(f, h.PrefixToNumber(0))
	fmt.Printf(f, h.PrefixToNumber(4))
	fmt.Printf(f, h.PrefixToNumber(8))
	fmt.Printf(f, h.PrefixToNumber(12))
	fmt.Printf(f, h.PrefixToNumber(16))
	fmt.Printf(f, h.PrefixToNumber(20))
	fmt.Printf(f, h.PrefixToNumber(24))
	fmt.Printf(f, h.PrefixToNumber(28))
	fmt.Printf(f, h.PrefixToNumber(32))
	// Output:
	// 0
	// 0
	// 0
	// 10
	// 100100
	// 1001000011
	// 10010000111111
	// 100100001111110110
	// 1001000011111101101010
	// 10010000111111011010101000
	// 100100001111110110101010001000
}
