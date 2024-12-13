package engine

const (
	InfinityCPValue int16 = 10_000
	DrawCPValue     int16 = 0

	PawnCPValue   int16 = 100
	KnightCPValue int16 = 300
	BishopCPValue int16 = 300
	RookCPValue   int16 = 500
	QueenCPValue  int16 = 950
)

var PieceValue = [5]int16{
	PawnCPValue,
	KnightCPValue,
	BishopCPValue,
	RookCPValue,
	QueenCPValue,
}

func evaluatePosition(pos *Position) int16 {
	scores := []int16{0, 0}
	evaluateMaterial(pos, White, scores)
	evaluateMaterial(pos, Black, scores)
	return scores[pos.Side] - scores[pos.Side^1]
}

func evaluateMaterial(pos *Position, side uint8, scores []int16) {
	piecesBB := pos.Colors[side] & ^pos.Pieces[King]
	for piecesBB != 0 {
		sq := GetLSBpos(piecesBB)
		scores[side] += PieceValue[pos.getPieceTypeOnSq(sq)]
		piecesBB &= (piecesBB - 1)
	}
}