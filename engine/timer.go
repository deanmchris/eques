package engine

import "time"

const (
	MovesToGoTimingFormat = iota
	SuddenDeathTimeFormat
	InfiniteTimeFormat
	NoFormat

	TimeBuffer          int64 = 100
	AvgExpectedNumMoves int64 = 70
	SmallestDivide      int64 = 8
)

type Timer struct {
	startTime       time.Time
	searchTime,
	movesToGo,
	movesToGoHalved,
	coeff           int64
	Stopped,
	infiniteTime    bool
}

func (timer *Timer) Init() {
	timer.movesToGo = AvgExpectedNumMoves
	timer.movesToGoHalved = timer.movesToGo / 2
	timer.coeff = (timer.movesToGoHalved * timer.movesToGoHalved) / 50
}

func (timer *Timer) CalculateSearchTime(timeFormat int, movesToGo, timeLeft, timeInc int64, numOfMoves uint16) {
	timer.Stopped = false
	bonus := timeInc / 2

	switch timeFormat {
	case MovesToGoTimingFormat:
		timer.searchTime = timeLeft / movesToGo + bonus
		timer.infiniteTime = false
	case SuddenDeathTimeFormat:
		timer.searchTime = timeLeft / timer.CalcTimeLeftDivide(numOfMoves) + bonus
		timer.infiniteTime = false
	case InfiniteTimeFormat:
		timer.infiniteTime = true
	}

	if timer.searchTime > TimeBuffer {
		timer.searchTime -= bonus
		if timer.searchTime > TimeBuffer {
			timer.searchTime -= TimeBuffer
		}
	}
}

func (timer *Timer) Start() {
	timer.startTime = time.Now()
}

func (timer *Timer) Update()  {
	if !timer.infiniteTime && time.Since(timer.startTime).Milliseconds() >= timer.searchTime {
		timer.Stopped = true
	}
}

func (timer *Timer) CalcTimeLeftDivide(numOfMoves uint16) int64 {
	numOfMovesInt64 := int64(numOfMoves)
	
	if numOfMovesInt64 <= timer.movesToGoHalved {
		return ((numOfMovesInt64 - timer.movesToGoHalved) * 
				(numOfMovesInt64 - timer.movesToGoHalved)) / timer.coeff + SmallestDivide
	}
	return (2 * (numOfMovesInt64 - timer.movesToGoHalved)) / timer.coeff + SmallestDivide
}
