package engine

import (
	"time"
)

const (
	MovesToGoTimingFormat = iota
	SuddenDeathTimeFormat
	InfiniteTimeFormat
	NoFormat

	TimeBuffer int64 = 150
)

type Timer struct {
	startTime    time.Time
	searchTime   int64
	Stopped      bool
	infiniteTime bool
}

func (timer *Timer) CalculateSearchTime(timeFormat int, movesToGo, timeLeft, timeInc int64) {
	timer.Stopped = false
	switch timeFormat {
	case MovesToGoTimingFormat:
		timer.searchTime = timeLeft / movesToGo + (timeInc / 2)
		timer.infiniteTime = false
	case SuddenDeathTimeFormat:
		timer.searchTime = timeLeft / 5 + (timeInc / 2)
		timer.infiniteTime = false
	case InfiniteTimeFormat:
		timer.infiniteTime = true
	}

	if timer.searchTime > TimeBuffer {
		timer.searchTime -= TimeBuffer
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
