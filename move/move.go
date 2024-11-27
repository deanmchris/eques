package move

const (
	QUIET uint8 = iota
	ATTACK
	WHT_ATTACK_EP
	BLK_ATTACK_EP
	PROMO_Q
	PROMO_R
	PROMO_B
	PROMO_N
	PROMO_ATTK_Q
	PROMO_ATTK_R
	PROMO_ATTK_B
	PROMO_ATTK_N
	WHT_CASTLE_K
	WHT_CASTLE_Q
	BLK_CASTLE_K
	BLK_CASTLE_Q

	FROM_SQ_BITMASK            = 0x3f
	TO_SQ_BITMASK              = 0xfc0
	MOVE_TYPE_BITMASK          = 0xf000
	MOVE_SCORE_BITMASK         = 0xffff0000
	FLIPPED_MOVE_SCORE_BITMASK = 0xffff
)

type Move uint32

func NewMove(fromSq, toSq, moveType uint8) Move {
	return Move(uint32(fromSq) | (uint32(toSq) << 6) | (uint32(moveType) << 12))
}

func (move Move) FromSq() uint8 {
	return uint8(move & FROM_SQ_BITMASK)
}

func (move Move) ToSq() uint8 {
	return uint8((move & TO_SQ_BITMASK) >> 6)
}

func (move Move) Type() uint8 {
	return uint8((move & MOVE_TYPE_BITMASK) >> 12)
}

func (move Move) Score() uint32 {
	return uint32((move & MOVE_SCORE_BITMASK) >> 16)
}

func (move *Move) SetScore(score uint32) {
	*move &= FLIPPED_MOVE_SCORE_BITMASK
	*move |= (Move(score) << 16)
}