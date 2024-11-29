package movegen

import (
	"bullet/position"
	"bullet/prng"
	"math/bits"
)

const (
	Rank1 uint8 = iota
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
)

const (
	FileA uint8 = iota
	FileB
	FileC
	FileD
	FileE
	FileF
	FileG
	FileH
)

const (
	North uint8 = 8
	South uint8 = 8
	East  uint8 = 1
	West  uint8 = 1
)

const (
	MaxBitsInRookBlockerMask = 4096
	MaxBitsInBishopBlockerMask = 512
)

// Optimized seeding values to find magic numbers based on file/rank, from Stockfish:
var MagicSeeds [8]uint64 = [8]uint64{728, 10316, 55013, 32803, 12281, 15100, 16645, 255}

type Magic struct {
	MagicNo      uint64
	BlockerMask uint64
	Shift        uint8
}

var RookMagics [64]Magic
var BishopMagics [64]Magic

var RookMovesHashTable [64][MaxBitsInRookBlockerMask]uint64
var BishopMovesHashTable [64][MaxBitsInBishopBlockerMask]uint64

var ClearRank = [8]uint64{}
var ClearFile = [8]uint64{}
var MaskRank = [8]uint64{}
var MaskFile = [8]uint64{}

var MaskDiagonal = [64]uint64{}
var MaskAntidiagonal = [64]uint64{}

var KingMoves = [64]uint64{}
var KnightMoves = [64]uint64{}
var PawnAttacks = [2][64]uint64{}
var PawnPushes = [2][64]uint64{}


func InitTables() {
	prng := prng.PseduoRandomGenerator{}

	genFileTables()
	genRankTables()

	for sq := uint8(0); sq < 64; sq++ {
		prng.Seed(MagicSeeds[position.RankOf(sq)])

		MaskDiagonal[sq] = genDiagonalRayGoingThruSq(sq)
		MaskAntidiagonal[sq] = genAntidiagonalRayGoingThruSq(sq)

		KingMoves[sq] = genKingMovesFromSq(sq)
		KnightMoves[sq] = genKnightMovesFromSq(sq)
		PawnPushes[position.White][sq] = genPawnPushMoveFromSq(sq, position.White)
		PawnPushes[position.Black][sq] = genPawnPushMoveFromSq(sq, position.Black)
		PawnAttacks[position.White][sq] = genPawnAttackMovesFromSq(sq, position.White)
		PawnAttacks[position.Black][sq] = genPawnAttackMovesFromSq(sq, position.Black)

		genRookMagicForSq(sq, &prng)
		genBishopMagicForSq(sq, &prng)
	}
}

func genFileTables() {
	for i := uint8(0); i < 8; i++ {
		emptyBB := position.EmptyBB
		fullBB := position.FullBB

		for j := i; j <= 63; j += 8 {
			emptyBB = position.SetBit(emptyBB, j)
			fullBB = position.UnsetBit(fullBB, j)
		}

		MaskFile[i] = emptyBB
		ClearFile[i] = fullBB
	}
}

func genRankTables() {
	for i := uint8(0); i <= 56; i += 8 {
		emptyBB := position.EmptyBB
		fullBB := position.FullBB

		for j := i; j < i+8; j++ {
			emptyBB = position.SetBit(emptyBB, j)
			fullBB = position.UnsetBit(fullBB, j)
		}

		MaskRank[i/8] = emptyBB
		ClearRank[i/8] = fullBB
	}
}

func genDiagonalRayGoingThruSq(sq uint8) uint64 {
	sqBBMaskFileHRank8 := MaskFile[FileH] | MaskRank[Rank8]
	sqBBMaskFileARank1 := MaskFile[FileA] | MaskRank[Rank1]

	sqBB := position.SquareBB[sq]
	diagonalMask := sqBB
		
	for i := uint8(1); (diagonalMask & sqBBMaskFileHRank8) == 0; i++ {
		diagonalMask |= sqBB << (i*North) << (i*East)
	}

	for i := uint8(1); (diagonalMask & sqBBMaskFileARank1) == 0; i++ {
		diagonalMask |= sqBB >> (i*South) >> (i*West)
	}
	return diagonalMask
}

func genAntidiagonalRayGoingThruSq(sq uint8) uint64 {
	sqBBMaskFileARank8 := MaskFile[FileA] | MaskRank[Rank8]
	sqBBMaskFileHRank1 := MaskFile[FileH] | MaskRank[Rank1]

	sqBB := position.SquareBB[sq]
	antidiagonalMask := sqBB

	for i := uint8(1); (antidiagonalMask & sqBBMaskFileARank8) == 0; i++ {
		antidiagonalMask |= sqBB << (i*North) >> (i*West)
	}

	for i := uint8(1); (antidiagonalMask & sqBBMaskFileHRank1) == 0; i++ {
		antidiagonalMask |= sqBB >> (i*South) << (i*East)
	}

	return antidiagonalMask
}

func genKingMovesFromSq(sq uint8) uint64 {
	sqBB := position.SquareBB[sq]
	sqBBClippedHFile := sqBB & ClearFile[FileH]
	sqBBClippedAFile := sqBB & ClearFile[FileA]

	top := sqBB << North
	topRight := sqBBClippedHFile << North << East
	topLeft := sqBBClippedAFile << North >> West

	right := sqBBClippedHFile << East
	left := sqBBClippedAFile >> West

	bottom := sqBB >> South
	bottomRight := sqBBClippedHFile >> South << East
	bottomLeft := sqBBClippedAFile >> South >> West

	return top | topRight | topLeft | right | left | bottom | bottomRight | bottomLeft
}

func genKnightMovesFromSq(sq uint8) uint64 {
	sqBB := position.SquareBB[sq]
	sqBBClippedHFile := sqBB & ClearFile[FileH]
	sqBBClippedAFile := sqBB & ClearFile[FileA]
	sqBBClippedHGFile := sqBB & ClearFile[FileH] & ClearFile[FileG]
	sqBBClippedABFile := sqBB & ClearFile[FileA] & ClearFile[FileB]

	northNorthEast := sqBBClippedHFile << North << North << East
	northEastEast := sqBBClippedHGFile << North << East << East

	southEastEast := sqBBClippedHGFile >> South << East << East
	southSouthEast := sqBBClippedHFile >> South >> South << East

	southSouthWest := sqBBClippedAFile >> South >> South >> West
	southWestWest := sqBBClippedABFile >> South >> West >> West

	northNorthWest := sqBBClippedAFile << North << North >> West
	northWestWest := sqBBClippedABFile << North >> West >> West

	return northNorthEast | northEastEast | southEastEast | southSouthEast |
		southSouthWest | southWestWest | northNorthWest | northWestWest
}

func genPawnPushMoveFromSq(sq, color uint8) uint64 {
	sqBB := position.SquareBB[sq]
	sqBBClipped18Rank := sqBB & ClearRank[Rank1] & ClearRank[Rank8]

	if color == position.White {
		return sqBBClipped18Rank << North
	}
	return sqBBClipped18Rank >> South
}

func genPawnAttackMovesFromSq(sq, color uint8) uint64 {
	sqBB := position.SquareBB[sq]
	sqBBClippedHFile := sqBB & ClearFile[FileH]
	sqBBClippedAFile := sqBB & ClearFile[FileA]
	sqBBClipped18Rank := sqBB & ClearRank[Rank1] & ClearRank[Rank8]

	if color == position.White {
		whitePawnRightAttack := (sqBBClippedHFile & sqBBClipped18Rank) << North << East
		whitePawnLeftAttack := (sqBBClippedAFile & sqBBClipped18Rank) << North >> West
		return  whitePawnRightAttack | whitePawnLeftAttack
	}

	blackPawnRightAttack := (sqBBClippedHFile & sqBBClipped18Rank) >> South << East
	blackPawnLeftAttack := (sqBBClippedAFile & sqBBClipped18Rank) >> South >> West
	return blackPawnRightAttack | blackPawnLeftAttack
}

func genRookMagicForSq(sq uint8, prng *prng.PseduoRandomGenerator) {
	magic := &RookMagics[sq]
	magic.BlockerMask = genRookMovesHQ(sq, position.EmptyBB, true)

	no_bits := bits.OnesCount64(magic.BlockerMask)
	magic.Shift = uint8(64 - no_bits)

	blockerMaskPermuations := make([]uint64, 1<<no_bits)
	blockerMaskPermutationMoves := make([]uint64, 1<<no_bits)

	blockers := position.EmptyBB
	index := 0

	for ok := true; ok; ok = (blockers != 0) {
		blockerMaskPermuations[index] = blockers
		blockerMaskPermutationMoves[index] = genRookMovesHQ(sq, blockers, false)

		index++
		blockers = (blockers - magic.BlockerMask) & magic.BlockerMask
	}

	searching := true
	possibleMagicNo := uint64(0)

	for searching {
		searching = false
		possibleMagicNo = prng.SparseRandom64()
		
		RookMovesHashTable[sq] = [MaxBitsInRookBlockerMask]uint64{}

		for i := 0; i < 1<<no_bits; i++ {
			blockerMaskPermutation := blockerMaskPermuations[i]
			hash := (blockerMaskPermutation * possibleMagicNo) >> magic.Shift 

			if RookMovesHashTable[sq][hash] != position.EmptyBB && 
			    RookMovesHashTable[sq][hash] != blockerMaskPermutationMoves[i] {
					searching = true
					break
			}

			RookMovesHashTable[sq][hash] = blockerMaskPermutationMoves[i]
		}
	}

	magic.MagicNo = possibleMagicNo
}

func genBishopMagicForSq(sq uint8, prng *prng.PseduoRandomGenerator) {
	magic := &BishopMagics[sq]
	magic.BlockerMask = genBishopMovesHQ(sq, position.EmptyBB, true)

	no_bits := bits.OnesCount64(magic.BlockerMask)
	magic.Shift = uint8(64 - no_bits)

	blockerMaskPermuations := make([]uint64, 1<<no_bits)
	blockerMaskPermutationMoves := make([]uint64, 1<<no_bits)

	blockers := position.EmptyBB
	index := 0

	for ok := true; ok; ok = (blockers != 0) {
		blockerMaskPermuations[index] = blockers
		blockerMaskPermutationMoves[index] = genBishopMovesHQ(sq, blockers, false)

		index++
		blockers = (blockers - magic.BlockerMask) & magic.BlockerMask
	}

	searching := true
	possibleMagicNo := uint64(0)

	for searching {
		searching = false
		possibleMagicNo = prng.SparseRandom64()
	
		BishopMovesHashTable[sq] = [MaxBitsInBishopBlockerMask]uint64{}

		for i := 0; i < 1<<no_bits; i++ {
			blockerMaskPermutation := blockerMaskPermuations[i]
			hash := (blockerMaskPermutation * possibleMagicNo) >> magic.Shift 

			if BishopMovesHashTable[sq][hash] != position.EmptyBB && 
			    BishopMovesHashTable[sq][hash] != blockerMaskPermutationMoves[i] {
					searching = true
					break
			}

			BishopMovesHashTable[sq][hash] = blockerMaskPermutationMoves[i]
		}
	}

	magic.MagicNo = possibleMagicNo
}


func genRookMovesHQ(sq uint8, occupiedBB uint64, genBlockerMask bool) uint64 {
	sliderBB := position.SquareBB[sq]

	fileMask := MaskFile[position.FileOf(sq)]
	rankMask := MaskRank[position.RankOf(sq)]

	rhs := bits.Reverse64(bits.Reverse64((occupiedBB & rankMask)) - (2 * bits.Reverse64(sliderBB)))
	lhs := (occupiedBB & rankMask) - 2*sliderBB
	eastWestMoves := (rhs ^ lhs) & rankMask

	rhs = bits.Reverse64(bits.Reverse64((occupiedBB & fileMask)) - (2 * bits.Reverse64(sliderBB)))
	lhs = (occupiedBB & fileMask) - 2*sliderBB
	northSouthMoves := (rhs ^ lhs) & fileMask

	if genBlockerMask {
		northSouthMoves &= ClearRank[Rank1] & ClearRank[Rank8]
		eastWestMoves &= ClearFile[FileA] & ClearFile[FileH]
	}

	return northSouthMoves | eastWestMoves
}

func genBishopMovesHQ(sq uint8, occupiedBB uint64, genBlockerMask bool) uint64 {
	sliderBB := position.SquareBB[sq]

	diagonalMask := MaskDiagonal[sq]
	antidiagonalMask := MaskAntidiagonal[sq]

	rhs := bits.Reverse64(bits.Reverse64((occupiedBB & diagonalMask)) - (2 * bits.Reverse64(sliderBB)))
	lhs := (occupiedBB & diagonalMask) - 2*sliderBB
	diagonalMoves := (rhs ^ lhs) & diagonalMask

	rhs = bits.Reverse64(bits.Reverse64((occupiedBB & antidiagonalMask)) - (2 * bits.Reverse64(sliderBB)))
	lhs = (occupiedBB & antidiagonalMask) - 2*sliderBB
	antidiagonalMoves := (rhs ^ lhs) & antidiagonalMask

	edges := position.FullBB
	if genBlockerMask {
		edges = ClearFile[FileA] & ClearFile[FileH] & ClearRank[Rank1] & ClearRank[Rank8]
	}
	return (diagonalMoves | antidiagonalMoves) & edges
}