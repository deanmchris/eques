package engine

import (
	"bullet/utils"
	"fmt"
	"strings"
)

const (
	MaxDepth               = 80
	NullMove          Move = 0
	LongestCheckmate int16 = 9000
)

type PVLine struct {
	moves [MaxDepth]Move
	cnt   uint8
}

func (pv *PVLine) bestMove() Move {
	return pv.moves[0]
}

func (pv *PVLine) setBestMove(move Move) {
	pv.moves[0] = move
}

func (pv *PVLine) extend(other *PVLine) {
	pv.cnt = 1
	for i := uint8(0); i < other.cnt; i++ {
		pv.moves[i+1] = other.moves[i]
		pv.cnt++
	}
}

func (pv *PVLine) clear() {
	pv.cnt = 0
}

func (pv *PVLine) String() string {
	sb := strings.Builder{}
	for i := uint8(0); i < pv.cnt; i++ {
		sb.WriteString(pv.moves[i].String())
		sb.WriteString(" ")
	}
	return sb.String()
}

type SearchData struct {
	posStack    [MaxDepth]Position
	pvLineStack [MaxDepth]PVLine
	Pos         Position
	totalNodes  uint64
}

func Search(sd *SearchData) Move {
	sd.totalNodes = 0
	bestMove := NullMove

	for depth := uint8(1); depth <= 10; depth++ {
		score := negamax(sd, -InfinityCPValue, InfinityCPValue, depth, 0)
		bestMove = sd.pvLineStack[0].bestMove()

		fmt.Printf(
			"info depth %d score cp %s nodes %d pv %s\n",
			depth,
			convertToUCIScore(score), 
			sd.totalNodes, 
			&sd.pvLineStack[0],
		)

		// info depth 2 score cp -38 time 22 nodes 132 nps 2659 pv e7e6 e2e4
	}

	return bestMove
}

func negamax(sd *SearchData, alpha, beta int16, depth, ply uint8) int16 {
	sd.totalNodes++

	if depth == 0 {
		return evaluatePosition(&sd.Pos)
	}

	noLegalMovesFlag := true

	for _, move := range genMoves(&sd.Pos) {
		CopyPos(&sd.Pos, &sd.posStack[ply])
		sd.Pos.DoMove(move)

		if sd.Pos.IsSideInCheck(sd.Pos.Side ^ 1) {
			CopyPos(&sd.posStack[ply], &sd.Pos)
			continue
		}

		noLegalMovesFlag = false
		score := -negamax(sd, -beta, -alpha, depth-1, ply+1)

		if score >= beta {
			return beta
		}

		if score > alpha {
			sd.pvLineStack[ply].clear()
			sd.pvLineStack[ply].setBestMove(move)
			sd.pvLineStack[ply].extend(&sd.pvLineStack[ply+1])
			alpha = score
		}

		CopyPos(&sd.posStack[ply], &sd.Pos)
	}

	if noLegalMovesFlag {
		sd.pvLineStack[ply].clear()
		if sd.Pos.IsSideInCheck(sd.Pos.Side) {
			return -InfinityCPValue + int16(ply)
		}
		return DrawCPValue
	}

	return alpha
}

func convertToUCIScore(score int16) string {
	score = utils.Abs(score)
	if score >= LongestCheckmate && score % 2 == 0 {
		return fmt.Sprintf("mate %d", (InfinityCPValue - score) / 2)
	}
	if score >= LongestCheckmate && score % 2 == 1 {
		return fmt.Sprintf("mate %d", (InfinityCPValue - score) / 2 + 1)
	}
	return fmt.Sprintf("%d", score)
}