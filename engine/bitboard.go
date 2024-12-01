package engine

import (
	"fmt"
	"math/bits"
)

// Bijective map chosen: A1 <-> LSB, B1 <-> LSB+1, . . ., G8 <-> MSB-1, H8 <-> MSB

const (
	FullBB uint64 = 0xffffffffffffffff
	EmptyBB uint64 = 0x0
)

func SetBit(bb uint64, sq uint8) uint64 {
	return bb | (1 << sq)
}

func UnsetBit(bb uint64, sq uint8) uint64 {
	return bb ^ (1 << sq)
}

func IsBitset(bb uint64, sq uint8) bool {
	return bb&(1 << sq) != 0
}

func GetLSBpos(bb uint64) uint8 {
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
