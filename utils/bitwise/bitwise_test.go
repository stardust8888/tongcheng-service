package bitwise

import (
	"fmt"
	"testing"
)

func TestBitwise(t *testing.T) {
	a := uint8(0b1010_0101) // 165
	b := uint8(0b0110_0011) // 99

	// 1. 位與 AND
	and := a & b // 0b0010_0001   同为1则1,否则为0
	fmt.Printf("a & b = %08b (%d)\n", and, and)

	// 2. 位或 OR
	or := a | b // 0b1110_0111    有1则1,否则为0
	fmt.Printf("a | b = %08b (%d)\n", or, or)

	// 3. 位異或 XOR
	xor := a ^ b // 0b1100_0110   不同则为1，相同则为0
	fmt.Printf("a ^ b = %08b (%d)\n", xor, xor)

	// 4. 位取反 NOT               为0则1，为1则0
	not := ^a // 一元運算子，結果為 uint8 最大值減 a
	fmt.Printf("^a   = %08b (%d)\n", not, not)

	// 5. 左移 LEFT SHIFT
	lsh := a << 2 // 0b1010_0101 → 0b1001_0100
	fmt.Printf("a<<2 = %08b (%d)\n", lsh, lsh)

	// 6. 右移 RIGHT SHIFT
	rsh := a >> 3 // 0b1010_0101 → 0b0001_0100
	fmt.Printf("a>>3 = %08b (%d)\n", rsh, rsh)
}
