package position

import (
	"fmt"
	"math/bits"
)

var SquareBB [64]uint64

func setBit(bb uint64, sq uint8) uint64 {
	return bb | SquareBB[sq]
}

func unsetBit(bb uint64, sq uint8) uint64 {
	return bb ^ SquareBB[sq]
}

func isBitset(bb uint64, sq uint8) bool {
	return bb&SquareBB[sq] != 0
}

func getLSBpos(bb uint64) uint8 {
	return uint8(bits.TrailingZeros64(bb))
}

func PrintBB(bb uint64) {
	bitstring := fmt.Sprintf("%064b", bb)
	for i := 7; i <= 63; i += 8 {
		fmt.Printf("%d |", (7-i/8)+1)
		for j := i; j >= i-7; j-- {
			bit := bitstring[j]
			if bit == '0' {
				fmt.Print(" .")
			} else {
				fmt.Print(" 1")
			}
		}
		fmt.Println()
	}
	fmt.Println("    ---------------\n    a b c d e f g h")
}

func init() {
	for i := 0; i < 64; i++ {
		// Bijective map chosen: A1 <-> LSB, B1 <-> LSB+1, . . ., G8 <-> MSB-1, H8 <-> MSB
		SquareBB[i] = 1 << i
	}
}
