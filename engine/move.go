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
	MoveTypeBitmask         = 0xf000
	MoveScoreBitmask        = 0xffff0000
	FlippedMoveScoreBitmask = 0xffff
)

type Move uint32

func NewMove(fromSq, toSq, moveType uint8) Move {
	return Move(uint32(fromSq) | (uint32(toSq) << 6) | (uint32(moveType) << 12))
}

func (move Move) FromSq() uint8 {
	return uint8(move & FromSqBitmask)
}

func (move Move) ToSq() uint8 {
	return uint8((move & ToSqBitmask) >> 6)
}

func (move Move) Type() uint8 {
	return uint8((move & MoveTypeBitmask) >> 12)
}

func (move Move) Score() uint32 {
	return uint32((move & MoveScoreBitmask) >> 16)
}

func (move *Move) SetScore(score uint32) {
	*move &= FlippedMoveScoreBitmask
	*move |= (Move(score) << 16)
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
