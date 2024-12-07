package engine

import "fmt"

const (
	StartingMoveListSize                  = 80
	DeltaToGenerateAttackPromotions uint8 = 4
	DeltaToGenerateQuietPromotions  uint8 = 0

	F1_G1_Mask    = 0x60
	B1_C1_D1_Mask = 0xe
	F8_G8_Mask    = 0x6000000000000000
	B8_C8_D8_Mask = 0xe00000000000000
)

func GenMoves(pos *Position) (moves []Move) {
	moves = make([]Move, 0, StartingMoveListSize)
	usBB := pos.Colors[pos.Side]
	enemyBB := pos.Colors[pos.Side^1]

	moves = genKnightMoves(pos, moves, usBB, enemyBB)
	moves = genBishopMoves(pos, moves, usBB, enemyBB)
	moves = genRookMoves(pos, moves, usBB, enemyBB)
	moves = genQueenMoves(pos, moves, usBB, enemyBB)
	moves = genNonCastlingKingMoves(pos, moves, usBB, enemyBB)

	if pos.Side == White {
		moves = genWhitePawnMoves(pos, moves, usBB, enemyBB)
		moves = genWhiteCastlingMoves(pos, moves, usBB, enemyBB)
	} else {
		moves = genBlackPawnMoves(pos, moves, usBB, enemyBB)
		moves = genBlackCastlingMoves(pos, moves, usBB, enemyBB)
	}

	return moves
}

func genWhitePawnMoves(pos *Position, moves []Move, usBB, enemyBB uint64) []Move {
	enemyBB |= (1 << pos.EPSq)
	pawnsBB := pos.Pieces[Pawn] & usBB

	pawnSinglePushMoves := (pawnsBB << North) & ^(usBB | enemyBB)
	pawnDoublePushMoves := ((pawnSinglePushMoves & MaskRank[Rank3]) << North) & ^(usBB | enemyBB)

	for pawnSinglePushMoves != 0 {
		to := GetLSBpos(pawnSinglePushMoves)
		pawnSinglePushMoves &= (pawnSinglePushMoves - 1)
		from := to - South

		if to >= A8 {
			moves = makePromotionMoves(from, to, DeltaToGenerateQuietPromotions, moves)
			continue
		}
		moves = append(moves, NewMove(from, to, Pawn, Quiet))
	}

	for pawnDoublePushMoves != 0 {
		to := GetLSBpos(pawnDoublePushMoves)
		from := to - South - South
		moves = append(moves, NewMove(from, to, Pawn, Quiet))
		pawnDoublePushMoves &= (pawnDoublePushMoves - 1)
	}

	pawnRightAttackMoves := ((pawnsBB & ClearFile[FileH]) << North << East) & enemyBB
	pawnLeftAttackMoves := ((pawnsBB & ClearFile[FileA]) << North >> West) & enemyBB

	for pawnRightAttackMoves != 0 {
		to := GetLSBpos(pawnRightAttackMoves)
		pawnRightAttackMoves &= (pawnRightAttackMoves - 1)
		from := to - South - West

		if to == pos.EPSq {
			moves = append(moves, NewMove(from, to, Pawn, WhiteAttackEP))
		} else {
			if to >= A8 {
				moves = makePromotionMoves(from, to, DeltaToGenerateAttackPromotions, moves)
				continue
			}
			moves = append(moves, NewMove(from, to, Pawn, Attack))
		}
	}

	for pawnLeftAttackMoves != 0 {
		to := GetLSBpos(pawnLeftAttackMoves)
		pawnLeftAttackMoves &= (pawnLeftAttackMoves - 1)
		from := to - South + East

		if to == pos.EPSq {
			moves = append(moves, NewMove(from, to, Pawn, WhiteAttackEP))
		} else {
			if to >= A8 {
				moves = makePromotionMoves(from, to, DeltaToGenerateAttackPromotions, moves)
				continue
			}
			moves = append(moves, NewMove(from, to, Pawn, Attack))
		}
	}

	return moves
}

func genBlackPawnMoves(pos *Position, moves []Move, usBB, enemyBB uint64) []Move {
	enemyBB |= 1 << pos.EPSq
	pawnsBB := pos.Pieces[Pawn] & usBB

	pawnSinglePushMoves := (pawnsBB >> South) & ^(usBB | enemyBB)
	pawnDoublePushMoves := ((pawnSinglePushMoves & MaskRank[Rank6]) >> South) & ^(usBB | enemyBB)

	for pawnSinglePushMoves != 0 {
		to := GetLSBpos(pawnSinglePushMoves)
		pawnSinglePushMoves &= (pawnSinglePushMoves - 1)
		from := to + North

		if to <= H1 {
			moves = makePromotionMoves(from, to, DeltaToGenerateQuietPromotions, moves)
			continue
		}
		moves = append(moves, NewMove(from, to, Pawn, Quiet))
	}

	for pawnDoublePushMoves != 0 {
		to := GetLSBpos(pawnDoublePushMoves)
		from := to + North + North
		moves = append(moves, NewMove(from, to, Pawn, Quiet))
		pawnDoublePushMoves &= (pawnDoublePushMoves - 1)
	}

	pawnRightAttackMoves := ((pawnsBB & ClearFile[FileH]) >> South << East) & enemyBB
	pawnLeftAttackMoves := ((pawnsBB & ClearFile[FileA]) >> South >> West) & enemyBB

	for pawnRightAttackMoves != 0 {
		to := GetLSBpos(pawnRightAttackMoves)
		pawnRightAttackMoves &= (pawnRightAttackMoves - 1)
		from := to + North - West

		if to == pos.EPSq {
			moves = append(moves, NewMove(from, to, Pawn, BlackAttackEP))
		} else {
			if to <= H1 {
				moves = makePromotionMoves(from, to, DeltaToGenerateAttackPromotions, moves)
				continue
			}
			moves = append(moves, NewMove(from, to, Pawn, Attack))
		}
	}

	for pawnLeftAttackMoves != 0 {
		to := GetLSBpos(pawnLeftAttackMoves)
		pawnLeftAttackMoves &= (pawnLeftAttackMoves - 1)
		from := to + North + East

		if to == pos.EPSq {
			moves = append(moves, NewMove(from, to, Pawn, BlackAttackEP))
		} else {
			if to <= H1 {
				moves = makePromotionMoves(from, to, DeltaToGenerateAttackPromotions, moves)
				continue
			}
			moves = append(moves, NewMove(from, to, Pawn, Attack))
		}
	}

	return moves
}

func genKnightMoves(pos *Position, moves []Move, usBB, enemyBB uint64) []Move {
	knightsBB := pos.Pieces[Knight] & usBB
	for knightsBB != 0 {
		sq := GetLSBpos(knightsBB)
		knightMoves := (KnightMoves[sq] & ^usBB)
		moves = genMovesFromBB(sq, Knight, knightMoves, enemyBB, moves)
		knightsBB &= (knightsBB - 1)
	}
	return moves
}

func genBishopMoves(pos *Position, moves []Move, usBB, enemyBB uint64) []Move {
	bishopsBB := pos.Pieces[Bishop] & usBB
	occuipiedBB := usBB | enemyBB

	for bishopsBB != 0 {
		sq := GetLSBpos(bishopsBB)
		bishopMoves := LookupBishopMoves(sq, occuipiedBB) & ^usBB
		moves = genMovesFromBB(sq, Bishop, bishopMoves, enemyBB, moves)
		bishopsBB &= (bishopsBB - 1)
	}

	return moves
}

func genRookMoves(pos *Position, moves []Move, usBB, enemyBB uint64) []Move {
	rooksBB := pos.Pieces[Rook] & usBB
	occuipiedBB := usBB | enemyBB

	for rooksBB != 0 {
		sq := GetLSBpos(rooksBB)
		rookMoves := LookupRookMoves(sq, occuipiedBB) & ^usBB
		moves = genMovesFromBB(sq, Rook, rookMoves, enemyBB, moves)
		rooksBB &= (rooksBB - 1)
	}

	return moves
}

func genQueenMoves(pos *Position, moves []Move, usBB, enemyBB uint64) []Move {
	queensBB := pos.Pieces[Queen] & usBB
	occuipiedBB := usBB | enemyBB

	for queensBB != 0 {
		sq := GetLSBpos(queensBB)

		bishopMoves := LookupBishopMoves(sq, occuipiedBB)
		rookMoves := LookupRookMoves(sq, occuipiedBB)
		queenMoves := (bishopMoves | rookMoves) & ^usBB

		moves = genMovesFromBB(sq, Queen, queenMoves, enemyBB, moves)
		queensBB &= (queensBB - 1)
	}

	return moves
}

func genNonCastlingKingMoves(pos *Position, moves []Move, usBB, enemyBB uint64) []Move {
	kingBB := pos.Pieces[King] & usBB
	sq := GetLSBpos(kingBB)
	kingMoves := KingMoves[sq] & ^usBB
	moves = genMovesFromBB(sq, King, kingMoves, enemyBB, moves)
	return moves
}

func genWhiteCastlingMoves(pos *Position, moves []Move, usBB, enemyBB uint64) []Move {
	allPieces := usBB | enemyBB

	if pos.Castling&WhiteKingsideRight == 0 {
		goto genQueensideCastlingMove
	}
	if allPieces&F1_G1_Mask != 0 {
		goto genQueensideCastlingMove
	}
	if pos.SqIsAttacked(pos.Side, E1) {
		goto genQueensideCastlingMove
	}
	if pos.SqIsAttacked(pos.Side, F1) {
		goto genQueensideCastlingMove
	}
	if pos.SqIsAttacked(pos.Side, G1) {
		goto genQueensideCastlingMove
	}

	moves = append(moves, NewMove(E1, G1, King, WhiteCastleK))

genQueensideCastlingMove:

	if pos.Castling&WhiteQueensideRight == 0 {
		goto Done
	}
	if allPieces&B1_C1_D1_Mask != 0 {
		goto Done
	}
	if pos.SqIsAttacked(pos.Side, E1) {
		goto Done
	}
	if pos.SqIsAttacked(pos.Side, D1) {
		goto Done
	}
	if pos.SqIsAttacked(pos.Side, C1) {
		goto Done
	}

	moves = append(moves, NewMove(E1, C1, King, WhiteCastleQ))

Done:
	return moves
}

func genBlackCastlingMoves(pos *Position, moves []Move, usBB, enemyBB uint64) []Move {
	allPieces := usBB | enemyBB

	if pos.Castling&BlackKingsideRight == 0 {
		goto genQueensideCastlingMove
	}
	if allPieces&F8_G8_Mask != 0 {
		goto genQueensideCastlingMove
	}
	if pos.SqIsAttacked(pos.Side, E8) {
		goto genQueensideCastlingMove
	}
	if pos.SqIsAttacked(pos.Side, F8) {
		goto genQueensideCastlingMove
	}
	if pos.SqIsAttacked(pos.Side, G8) {
		goto genQueensideCastlingMove
	}

	moves = append(moves, NewMove(E8, G8, King, BlackCastleK))

genQueensideCastlingMove:

	if pos.Castling&BlackQueensideRight == 0 {
		goto Done
	}
	if allPieces&B8_C8_D8_Mask != 0 {
		goto Done
	}
	if pos.SqIsAttacked(pos.Side, E8) {
		goto Done
	}
	if pos.SqIsAttacked(pos.Side, D8) {
		goto Done
	}
	if pos.SqIsAttacked(pos.Side, C8) {
		goto Done
	}

	moves = append(moves, NewMove(E8, C8, King, BlackCastleQ))

Done:
	return moves
}

func genMovesFromBB(from, fromType uint8, movesBB, enemyBB uint64, moves []Move) []Move {
	for movesBB != 0 {
		to := GetLSBpos(movesBB)
		toBB := uint64(1) << to

		moveType := Quiet
		if toBB&enemyBB != 0 {
			moveType = Attack
		}

		moves = append(moves, NewMove(from, to, fromType, moveType))
		movesBB &= (movesBB - 1)
	}
	return moves
}

func makePromotionMoves(from, to uint8, deltaToGenQuietOrAttackPromos uint8, moves []Move) []Move {
	moves = append(moves, NewMove(from, to, Pawn, deltaToGenQuietOrAttackPromos+PromoQ))
	moves = append(moves, NewMove(from, to, Pawn, deltaToGenQuietOrAttackPromos+PromoR))
	moves = append(moves, NewMove(from, to, Pawn, deltaToGenQuietOrAttackPromos+PromoB))
	moves = append(moves, NewMove(from, to, Pawn, deltaToGenQuietOrAttackPromos+PromoN))
	return moves
}

func Perft(pos *Position, depth uint8, tt *TranspositionTable[PerftEntry]) uint64 {
	if depth == 0 {
		return 1
	}

	if tt.size > 0 {
		if entry := tt.Probe(pos.Hash); entry != nil && entry.Depth() == depth {
			return entry.Nodes()
		}
	}

	moves := GenMoves(pos)
	nodes := uint64(0)

	for _, move := range moves {
		newPos := pos.DoMove(move)
		if !newPos.IsSideInCheck(newPos.Side ^ 1) {
			nodes += Perft(newPos, depth-1, tt)
		}

	}

	if tt.size > 0 {
		tt.Store(pos.Hash, depth).SetData(pos.Hash, nodes, depth)
	}

	return nodes
}

func DPerft(pos *Position, depth uint8, tt *TranspositionTable[PerftEntry]) uint64 {
	var helper func(*Position, uint8, uint8, *TranspositionTable[PerftEntry]) uint64

	helper = func(pos *Position, depth, startDepth uint8, tt *TranspositionTable[PerftEntry]) uint64 {
		if depth == 0 {
			return 1
		}

		if tt.size > 0 {
			if entry := tt.Probe(pos.Hash); entry != nil && entry.Depth() == depth {
				return entry.Nodes()
			}
		}

		moves := GenMoves(pos)
		nodes := uint64(0)

		for _, move := range moves {
			newPos := pos.DoMove(move)
			if !newPos.IsSideInCheck(newPos.Side ^ 1) {
				moveNodes := helper(newPos, depth-1, depth, tt)
				if depth == startDepth {
					fmt.Printf("%v: %v\n", move, moveNodes)
				}
				nodes += moveNodes
			}
		}

		if tt.size > 0 {
			tt.Store(pos.Hash, depth).SetData(pos.Hash, nodes, depth)
		}

		return nodes
	}

	return helper(pos, depth, depth, tt)
}
