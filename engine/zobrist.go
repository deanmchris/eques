package engine

import (
	"bullet/prng"
	"time"
)

var PieceZobristValues [2][6][64]uint64
var EPSqZobristValues [65]uint64
var CastlingZobristValues [16]uint64
var SideZobristValues [2]uint64

func InitZobristValues() {
	prng := prng.PseduoRandomGenerator{}
	prng.Seed(uint64(time.Now().UnixNano()))

	for color := White; color <= Black; color++ {
		for piece := Pawn; piece <= King; piece++ {
			for sq := 0; sq < 64; sq++ {
				PieceZobristValues[color][piece][sq] = prng.Random64()
			}
		}
	}

	// Make sure the index corresponding to an invalid sq, i.e NoSq,
	// is 0, since we only want to XOR a non-zero into the hash for the 
	// possible en passant square if there is one.
	for sq := 0; sq < 64; sq++ {
		EPSqZobristValues[sq] = prng.Random64()
	}

	for i := 0; i < 16; i++ {
		CastlingZobristValues[i] = prng.Random64()
	}

	// Make sure the index corresponding to white is 0, since we only want
	// to XOR a non-zero number into the hash if the side to move is black.
	SideZobristValues[Black] = prng.Random64()
}

func GenHash(pos *Position) (hash uint64) {
	for sq := uint8(0); sq < 64; sq++ {
		pieceType := pos.GetPieceTypeOnSq(sq)
		pieceColor := pos.getPieceColorOnSq(sq)

		if pieceType == NoType {
			continue
		}

		hash ^= PieceZobristValues[pieceColor][pieceType][sq]
	}

	hash ^= EPSqZobristValues[pos.EPSq]
	hash ^= CastlingZobristValues[pos.Castling]
	hash ^= SideZobristValues[pos.Side]

	return hash
}