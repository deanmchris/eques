package tuner

import (
	"bullet/engine"
	"math/rand"
)

const (
	NumPSQTWeights = 6 * 64

	RandomDeltaBound float64 = 25

	BasePawnCPValue float64   = 100
	BaseKnightCPValue float64 = 300
	BaseBishopCPValue float64 = 300
	BaseRookCPValue float64   = 500
	BaseQueenCPValue float64  = 850
	BaseKingCPValue float64   = 0
)

var BasePieceValues = [6]float64{
	BasePawnCPValue,
	BaseBishopCPValue,
	BaseKnightCPValue,
	BaseRookCPValue,
	BaseQueenCPValue,
	BaseKingCPValue,
}

func genRandIntWithinSymmetricInterval(bound float64) float64 {
	return (rand.Float64() * 2*bound) - bound
}

type Weights struct {
	PSQTWeights [NumPSQTWeights]float64
}

func (weights *Weights) Randomize() {
	for pieceType := engine.Pawn; pieceType < engine.NoType; pieceType++ {
		baseValue := BasePieceValues[pieceType]
		startIdx := pieceType*64

		for sq := 0; sq < 64; sq++ {
			value := baseValue + genRandIntWithinSymmetricInterval(RandomDeltaBound)
			weights.PSQTWeights[startIdx+sq] = value
		}
	}
}

func (weights *Weights) LoadWeights(PSQT [6][64]int16) {
	for pieceType := engine.Pawn; pieceType < engine.NoType; pieceType++ {
		startIdx := pieceType*64
		for sq := 0; sq < 64; sq++ {
			weights.PSQTWeights[startIdx+sq] = float64(PSQT[pieceType][sq])
		}
	}
}

func (weights *Weights) CopyWeightsToPSQT(PSQT *[6][64]int16) {
	for pieceType := engine.Pawn; pieceType < engine.NoType; pieceType++ {
		startIdx := pieceType*64
		for sq := 0; sq < 64; sq++ {
			PSQT[pieceType][sq] = int16(weights.PSQTWeights[startIdx+sq])
		}
	}
}

type Piece struct {
	WeightIdx uint16
}

type PositionData struct {
	Pieces [2][]Piece
}

func (pd *PositionData) LoadData(pos *engine.Position) {
	piecesBB := pos.Colors[engine.White] | pos.Colors[engine.Black]

	for piecesBB != 0 {
		sq := engine.GetLSBpos(piecesBB)
		pieceType := pos.GetPieceTypeOnSq(sq)
		pieceColor := pos.GetPieceColorOnSq(sq)
		weightIdx := uint16(pieceType)*64 + uint16(engine.FlipSq[pieceColor][sq])

		pd.Pieces[pieceColor] = append(
			pd.Pieces[pieceColor], 
			Piece{WeightIdx: weightIdx},
		)
		piecesBB &= (piecesBB - 1)
	}
}

func evaluatePosition(weights *Weights, pd *PositionData) float64 {
	score := float64(0)
	for _, piece := range pd.Pieces[engine.White] {
		score += weights.PSQTWeights[piece.WeightIdx]
	}
	for _, piece := range pd.Pieces[engine.Black] {
		score -= weights.PSQTWeights[piece.WeightIdx]
	}
	return score
}