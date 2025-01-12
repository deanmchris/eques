package engine

import (
	"eques/utils"
	"fmt"
	"strings"
	"time"
)

const (
	MaxDepth               = 60
	MaxPly                 = 80
	MaxGameLength          = 1024
	NullMove          Move = 0
	LongestCheckmate int16 = 9000
	BestMoveScore   uint16 = 8000
)

var MVV_LVA [7][6]uint16 = [7][6]uint16{
	{15, 14, 13, 12, 11, 10}, // victim is pawn
	{25, 24, 23, 22, 21, 20}, // victim is knight
	{35, 34, 33, 32, 31, 30}, // victim is bishop
	{45, 44, 43, 42, 41, 40}, // victim is rook
	{55, 54, 53, 52, 51, 50}, // victim is queen
	{0,  0,  0,  0,  0,  0},  // victim is King
	{0,  0,  0,  0,  0,  0},  // victim is no piece
}

type PVLine struct {
	Moves [MaxPly]Move
	Cnt   uint8
}

func (pv *PVLine) bestMove() Move {
	return pv.Moves[0]
}

func (pv *PVLine) update(move Move, other *PVLine) {
	pv.Moves[0] = move
	pv.Cnt = 1
	for i := uint8(0); i < other.Cnt; i++ {
		pv.Moves[i+1] = other.Moves[i]
		pv.Cnt++
	}
}

func (pv *PVLine) clear() {
	pv.Cnt = 0
}

func (pv *PVLine) copy(other *PVLine) {
	pv.Cnt = 0
	for i := uint8(0); i < other.Cnt; i++ {
		pv.Moves[i] = other.Moves[i]
		pv.Cnt++
	}
}

func (pv *PVLine) String() string {
	sb := strings.Builder{}
	for i := uint8(0); i < pv.Cnt; i++ {
		sb.WriteString(pv.Moves[i].String())
		sb.WriteString(" ")
	}
	return sb.String()
}

type SearchData struct {
	posStack    [MaxPly]Position
	pvLineStack [MaxPly]PVLine
	posHistory  [MaxGameLength]uint64
	Timer       Timer
	Pos         Position
	prevPV      PVLine
	totalNodes  uint64
	historyIdx  uint16
}

func (sd *SearchData) Reset() {
	sd.posStack = [MaxPly]Position{}
	sd.pvLineStack = [MaxPly]PVLine{}
	sd.Pos = Position{}
	sd.posHistory = [MaxGameLength]uint64{}
	sd.historyIdx = 0
}

func (sd *SearchData) AddCurrPosToHistory() {
	sd.posHistory[sd.historyIdx] = sd.Pos.Hash
	sd.historyIdx++
}

func (sd *SearchData) PopFromPosHistory() {
	sd.historyIdx--
}

func (sd *SearchData) ClearPosHistory() {
	sd.historyIdx = 0
}

func (sd *SearchData) GetCurrPV() PVLine {
	return sd.pvLineStack[0]
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
			"info depth %d time %d score %s nodes %d pv %snps %d\n",
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

	if ply >= MaxPly {
		return EvaluatePosition(&sd.Pos)
	}

	sd.pvLineStack[ply].clear()

	isRoot := ply == 0
	inCheck := sd.Pos.IsSideInCheck(sd.Pos.Side)

	if !isRoot && nodeIsDraw(sd) {
		return DrawCPValue
	}

	if depth <= 2 && inCheck {
		depth++
	}
 
	if depth == 0 {
		return Qsearch(sd, alpha, beta, ply)
	}

	sd.totalNodes++
	noLegalMovesFlag := true

	moves := genMoves(&sd.Pos)
	scoreMoves(sd, moves, sd.prevPV.Moves[ply])
	moveOrderer := createMoveOrderer(moves)

	for move := moveOrderer(); move != NullMove; move = moveOrderer() {
		CopyPos(&sd.Pos, &sd.posStack[ply])
		sd.Pos.DoMove(move)

		if sd.Pos.IsSideInCheck(sd.Pos.Side ^ 1) {
			CopyPos(&sd.posStack[ply], &sd.Pos)
			continue
		}

		sd.AddCurrPosToHistory()

		noLegalMovesFlag = false
		score := -negamax(sd, -beta, -alpha, depth-1, ply+1)

		CopyPos(&sd.posStack[ply], &sd.Pos)
		sd.PopFromPosHistory()

		if score >= beta {
			return beta
		}

		if score > alpha {
			sd.pvLineStack[ply].update(move, &sd.pvLineStack[ply+1])
			alpha = score
		}
	}

	if noLegalMovesFlag {
		if inCheck {
			return -InfinityCPValue + int16(ply)
		}
		return DrawCPValue
	}

	return alpha
}

func Qsearch(sd *SearchData, alpha, beta int16, ply uint8) int16 {
	if sd.totalNodes & 2047 == 0 {
		sd.Timer.Update()
	}

	if sd.Timer.Stopped {
		return 0
	}

	if ply >= MaxPly {
		return EvaluatePosition(&sd.Pos)
	}

	sd.totalNodes++

	sd.pvLineStack[ply].clear()
	eval := EvaluatePosition(&sd.Pos)

	if eval >= beta {
		return beta
	}

	if eval > alpha {
		alpha = eval
	}

	moves := genAttacksAndQueenPromos(&sd.Pos)
	scoreMoves(sd, moves, sd.prevPV.Moves[ply])
	moveOrderer := createMoveOrderer(moves)

	for move := moveOrderer(); move != NullMove; move = moveOrderer() {
		CopyPos(&sd.Pos, &sd.posStack[ply])
		sd.Pos.DoMove(move)

		if sd.Pos.IsSideInCheck(sd.Pos.Side ^ 1) {
			CopyPos(&sd.posStack[ply], &sd.Pos)
			continue
		}

		score := -Qsearch(sd, -beta, -alpha, ply+1)

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
			mvv_lva_score := MVV_LVA[sd.Pos.GetPieceTypeOnSq(move.ToSq())][move.FromType()]
			move.SetScore(mvv_lva_score)
		}
	}
}

func nodeIsDraw(sd *SearchData) bool {
	if sd.Pos.HalfMove >= 100 {
		return true
	}

	for i := uint16(0); i < sd.historyIdx-1; i++ {
		if sd.posHistory[i] == sd.Pos.Hash {
			return true
		}
	}

	return false
}

func convertToUCIScore(score int16) string {
	scoreAbs := utils.Abs(score)
	if scoreAbs >= LongestCheckmate && scoreAbs % 2 == 0 {
		return fmt.Sprintf("mate %d", (InfinityCPValue - scoreAbs) / 2)
	}
	if scoreAbs >= LongestCheckmate && scoreAbs % 2 == 1 {
		return fmt.Sprintf("mate %d", (InfinityCPValue - scoreAbs) / 2 + 1)
	}
	return fmt.Sprintf("cp %d", score)
}