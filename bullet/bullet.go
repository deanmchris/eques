package main

import (
	"bullet/engine"
	"runtime/debug"
)

const (
	GC_TARGET_PERCENTAGE = 300
)

func init() {
	engine.InitTables()
	engine.InitZobristValues()
}

func main() {
	// Essentially setting this argument value higher makes Go's garbage collector less agressive,
	// which can improve the overall performance of the engine. This does come at the expense of 
	// higher memory usage when the engine is running. The current value seems to be a good balance
	// between these considerations.
	debug.SetGCPercent(GC_TARGET_PERCENTAGE)
}
