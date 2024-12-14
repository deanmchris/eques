package engine

import (
	"bullet/utils"
	"fmt"
	"strings"
	"time"
)

const (
	MaxDepth               = 60
	MaxPly                 = 80
	NullMove          Move = 0
	LongestCheckmate int16 = 9000
	BestMoveScore   uint16 = 8000
)

var MVV_LVA = [6][7]uint16{
	{52, 54, 56, 58, 60, 0, 0}, // pawn attacker
	{42, 44, 46, 48, 50, 0, 0}, // knight attacker
	{32, 34, 36, 38, 40, 0, 0}, // bishop attacker
	{22, 24, 26, 28, 30, 0, 0}, // rook attacker
	{12, 14, 16, 18, 20, 0, 0}, // queen attacker
	{2,   4,  6,  8, 10, 0, 0}, // king attacker
}

type PVLine struct {
	moves [MaxPly]Move
	cnt   uint8
}

func (pv *PVLine) bestMove() Move {
	return pv.moves[0]
}

func (pv *PVLine) update(move Move, other *PVLine) {
	pv.moves[0] = move
	pv.cnt = 1
	for i := uint8(0); i < other.cnt; i++ {
		pv.moves[i+1] = other.moves[i]
		pv.cnt++
	}
}

func (pv *PVLine) clear() {
	pv.cnt = 0
}

func (pv *PVLine) copy(other *PVLine) {
	pv.cnt = 0
	for i := uint8(0); i < other.cnt; i++ {
		pv.moves[i] = other.moves[i]
		pv.cnt++
	}
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
	posStack    [MaxPly]Position
	pvLineStack [MaxPly]PVLine
	Timer       Timer
	Pos         Position
	prevPV      PVLine
	totalNodes  uint64
}

func Search(sd *SearchData) Move {
	sd.totalNodes = 0
	sd.prevPV.clear()

	bestMove := NullMove
	sd.Timer.Start()
	totalTime := int64(0)

	for depth := uint8(1); depth <= MaxDepth; depth++ {
		startTime := time.Now()
		score := negamax(sd, -InfinityCPValue, InfinityCPValue, depth, 0)
		endTime := time.Since(startTime)

		if sd.Timer.Stopped {
			break
		}
		
		bestMove = sd.pvLineStack[0].bestMove()
		totalTime += endTime.Milliseconds()
		nps := (sd.totalNodes * 1000) / uint64(totalTime+1)

		fmt.Printf(
			"info depth %d time %d score cp %s nodes %d pv %snps %d\n",
			depth,
			totalTime,
			convertToUCIScore(score), 
			sd.totalNodes, 
			&sd.pvLineStack[0],
			nps,
		)
		
		sd.prevPV.copy(&sd.pvLineStack[0])
	}

	return bestMove
}

func negamax(sd *SearchData, alpha, beta int16, depth, ply uint8) int16 {
	if sd.totalNodes & 2047 == 0 {
		sd.Timer.Update()
	}

	if sd.Timer.Stopped {
		return 0
	}

	if depth == 0 {
		return qsearch(sd, alpha, beta, ply)
	}

	sd.totalNodes++

	sd.pvLineStack[ply].clear()
	noLegalMovesFlag := true

	moves := genMoves(&sd.Pos)
	scoreMoves(sd, moves, sd.prevPV.moves[ply])
	moveOrderer := createMoveOrderer(moves)

	for move := moveOrderer(); move != NullMove; move = moveOrderer() {
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
			sd.pvLineStack[ply].update(move, &sd.pvLineStack[ply+1])
			alpha = score
		}

		CopyPos(&sd.posStack[ply], &sd.Pos)
	}

	if noLegalMovesFlag {
		if sd.Pos.IsSideInCheck(sd.Pos.Side) {
			return -InfinityCPValue + int16(ply)
		}
		return DrawCPValue
	}

	return alpha
}

func qsearch(sd *SearchData, alpha, beta int16, ply uint8) int16 {
	if sd.totalNodes & 2047 == 0 {
		sd.Timer.Update()
	}

	if sd.Timer.Stopped {
		return 0
	}

	sd.totalNodes++

	sd.pvLineStack[ply].clear()
	eval := evaluatePosition(&sd.Pos)

	if eval >= beta {
		return beta
	}

	if eval > alpha {
		alpha = eval
	}

	moves := genAttacksAndQueenPromos(&sd.Pos)
	scoreMoves(sd, moves, sd.prevPV.moves[ply])
	moveOrderer := createMoveOrderer(moves)

	for move := moveOrderer(); move != NullMove; move = moveOrderer() {
		CopyPos(&sd.Pos, &sd.posStack[ply])
		sd.Pos.DoMove(move)

		if sd.Pos.IsSideInCheck(sd.Pos.Side ^ 1) {
			CopyPos(&sd.posStack[ply], &sd.Pos)
			continue
		}

		score := -qsearch(sd, -beta, -alpha, ply+1)

		if score >= beta {
			return beta
		}

		if score > alpha {
			sd.pvLineStack[ply].update(move, &sd.pvLineStack[ply+1])
			alpha = score
		}

		CopyPos(&sd.posStack[ply], &sd.Pos)
	}

	return alpha
}

func createMoveOrderer(moves []Move) func() Move {
	startedOfUnsortedSublist := 0
	return func() Move {
		if startedOfUnsortedSublist == len(moves) {
			return NullMove
		}

		bestMove := moves[startedOfUnsortedSublist]
		bestMoveIdx := startedOfUnsortedSublist

		for i := startedOfUnsortedSublist+1; i < len(moves); i++ {
			move := moves[i]
			if move.Score() > bestMove.Score() {
				bestMove = move
				bestMoveIdx = i
			}
		}

		tmp := moves[startedOfUnsortedSublist]
		moves[startedOfUnsortedSublist] = bestMove
		moves[bestMoveIdx] = tmp
		startedOfUnsortedSublist++

		return bestMove
	}
}

func scoreMoves(sd *SearchData, moves []Move, bestMoveFromPrevDepth Move) {
	for i := 0; i < len(moves); i++ {
		move := &moves[i]
		if move.Equal(bestMoveFromPrevDepth) {
			move.SetScore(BestMoveScore)
		} else {
			mvv_lva_score := MVV_LVA[move.FromType()][sd.Pos.getPieceTypeOnSq(move.ToSq())]
			move.SetScore(mvv_lva_score)
		}
	}
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