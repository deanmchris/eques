package main

import (
	"eques/engine"
	"eques/uci"
	"eques/tuner"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

const (
	GC_TARGET_PERCENTAGE = 300

	DefaultLearningRate float64 = 0.8
	DefaultIterations   int     = 2000
	DefaultRecordRate   int     = 40
	DefaultDepth        uint    = 1
	DefaultTTSize       uint64  = 16
	DefaultNumThreads   int     = 1
)

func init() {
	engine.InitTables()
	engine.InitZobristValues()
}

func processTuneCommand() {
	tuneCmd := flag.NewFlagSet("tune", flag.ExitOnError)
	tuneDataFile := tuneCmd.String(
		"infile", 
		"", 
		"The input file to the tuner. Should be a CSV file of fens in the first column, and the\n" +
		"outcome of the game in the second column (white win=1.0, black win=0.0, draw=0.5).",
	)

	tuneLearningRate := tuneCmd.Float64(
		"learning_rate",
		DefaultLearningRate,
		"The default learning rate to use when performing (AdaGrad) gradient descent.",
	)

	tuneIterations := tuneCmd.Int(
		"iterations",
		DefaultIterations,
		"The number of iterations to perform gradient descent.",
	)

	tuneNumThreads := tuneCmd.Int(
		"num_threads",
		DefaultNumThreads,
		"The number of \"threads\" (go-routines) to spawn to paralleize the tuning process.",
	)

	tuneRecordErrEveryNth := tuneCmd.Int(
		"record_err_every_nth",
		DefaultRecordRate,
		"Record the mean-square error every <record-err-every-nth> iterations. Used in conjuction\n" +
		"with the visualize_error_rate.py script to display a graph of the error rate over the course\n" +
		"of the tuning process.",
	)

	tuneCmd.Parse(os.Args[2:])

	if *tuneDataFile == "" {
		fmt.Println("Please supply a data file to the tuner.")
		return
	}

	weights := tuner.Weights{}
	weights.LoadBaseWeights()
	weights.TuneWeights(*tuneDataFile, *tuneLearningRate, *tuneIterations, *tuneNumThreads, *tuneRecordErrEveryNth)
}

func processPerftCommand() {
	perftCmd := flag.NewFlagSet("perft", flag.ExitOnError)

	perftFEN := perftCmd.String(
		"fen", 
		engine.FENStartPosition,
		"The position to run perft on as FEN string.",
	)

	perftDepth := perftCmd.Uint(
		"depth",
		DefaultDepth,
		"The depth to run perft to.",
	)

	perftVerbose := perftCmd.Bool(
		"verbose",
		false,
		"Print each legal move for the current position and the number of subnodes that result\n" +
		"from said move.",
	)

	perftTTSize := perftCmd.Uint64(
		"tt_size",
		DefaultTTSize,
		"The size to make the transposition table, in megabytes.",
	)

	perftCmd.Parse(os.Args[2:])

	pd := engine.PerftData{}
	pd.TT.SetSize(*perftTTSize, engine.PerftEntrySize)
	pd.Pos.LoadFEN(*perftFEN)

	depth := uint8(*perftDepth)
	var nodes uint64
	var startTime time.Time
	var endTime time.Duration
	
	if *perftVerbose {
		startTime = time.Now()
		nodes = engine.DPerft(&pd, depth, depth, 0)
		endTime = time.Since(startTime)
	} else {
		startTime = time.Now()
		nodes = engine.Perft(&pd, depth, 0)
		endTime = time.Since(startTime)
	}

	fmt.Println("nodes:", nodes)
	fmt.Printf("time: %d ms\n", endTime.Milliseconds())
	fmt.Printf("nps: %d\n", uint64(float64(nodes) / float64(endTime.Seconds())))
}

func main() {
	// Essentially setting this argument value higher makes Go's garbage collector less agressive,
	// which can improve the overall performance of the engine. This does come at the expense of 
	// higher memory usage when the engine is running. The current value seems to be a good balance
	// between these considerations.
	debug.SetGCPercent(GC_TARGET_PERCENTAGE)

	if len(os.Args) < 2 {
		uci.StartUCIProtocolInterface()
		return
	}

	switch os.Args[1] {
	case "tune":
		processTuneCommand()
	case "perft":
		processPerftCommand()
	case "uci":
		uci.StartUCIProtocolInterface()	
	case "-h", "h", "--help", "help":
		fmt.Println("The following commands are available:")
		fmt.Print(
			"    * tune: Run the tuner. Run the program with the flags \"tune -h\"\n" +
			"      for more details.\n",
			"    * tune: Run perft. Run the program with the flags \"perft -h\"\n" +
			"      for more details.\n",
			"    * uci: Start the UCI protocol. Program will default to this command if\n" +
			"      no command is given.\n",
		)
	default:
		fmt.Printf("unrecognized command-line argument: \"%s\"", os.Args[1])
	}
}
