package engine

import (
	"fmt"
)

const (
	Quiet uint8 = iota
	Attack
	WhiteAttackEP
	BlackAttackEP
	PromoQ
	PromoR
	PromoB
	PromoN
	PromoAttkQ
	PromoAttkR
	PromoAttkB
	PromoAttkN
	WhiteCastleK
	WhiteCastleQ
	BlackCastleK
	BlackCastleQ

	FromSqBitmask           = 0x3f
	ToSqBitmask             = 0xfc0
	FromTypeBitmask         = 0x7000
	MoveTypeBitmask         = 0x78000
	MoveScoreBitmask        = 0xfff80000
	FlippedMoveScoreBitmask = 0x7ffff
)

// A move is encoded as a 32 bit integer with the following structure (starting with LSB):
// 6-bits: from square
// 6-bits: to square
// 3-bits: piece type on from sq (color should always be the side to move)
// 4-bits: moveType
// 13-bits: move score
type Move uint32

func NewMove(fromSq, toSq, fromType, moveType uint8) Move {
	return Move(
		uint32(fromSq) | 
		(uint32(toSq) << 6) | 
		(uint32(fromType) << 12) |
		(uint32(moveType) << 15))
}

func (move Move) FromSq() uint8 {
	return uint8(move & FromSqBitmask)
}

func (move Move) ToSq() uint8 {
	return uint8((move & ToSqBitmask) >> 6)
}

func (move Move) FromType() uint8 {
	return uint8((move & FromTypeBitmask) >> 12)
}

func (move Move) Type() uint8 {
	return uint8((move & MoveTypeBitmask) >> 15)
}

func (move Move) Score() uint16 {
	return uint16((move & MoveScoreBitmask) >> 19)
}

func (move *Move) SetScore(score uint16) {
	*move &= FlippedMoveScoreBitmask
	*move |= (Move(score) << 19)
}

func (move Move) Equal(other Move) bool {
	return (move & FlippedMoveScoreBitmask) == (other & FlippedMoveScoreBitmask)
}

func (move Move) String() string {
	from, to, moveType := move.FromSq(), move.ToSq(), move.Type()

	promotionType := ""
	switch moveType {
	case PromoN, PromoAttkN:
		promotionType = "n"
	case PromoB, PromoAttkB:
		promotionType = "b"
	case PromoR, PromoAttkR:
		promotionType = "r"
	case PromoQ, PromoAttkQ:
		promotionType = "q"
	}
	return fmt.Sprintf("%v%v%v", SqToCoord(from), SqToCoord(to), promotionType)
}
