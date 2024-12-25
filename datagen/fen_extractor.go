package datagen

import (
	"bufio"
	"bullet/engine"
	"bullet/utils"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
)

const (
	BasePawnCP uint16       = 100
	BaseKnightCP uint16     = 320
	BaseBishopCP uint16     = 320
	BaseRookCP uint16       = 500
	BaseQueenCP uint16      = 900
	BaseKingCP uint16       = 0
	MaxMaterialCount uint16 = 16*BasePawnCP + 
	                          4*BaseKnightCP + 
							  4*BaseBishopCP + 
							  4*BaseRookCP + 
							  2*BaseQueenCP

	ReportEveryNGames = 1000
)


var BaseMaterialCP = [6]uint16 {
	BasePawnCP,
	BaseKnightCP,
	BaseBishopCP,
	BaseRookCP,
	BaseQueenCP,
	BaseKingCP,
}



func ExtractFENs(pgnFilePath, outFilePath string, sampleSizePerGame uint16, scoreBoundCP int16) {
	parser := PGNParser{}
	parser.LoadPGNFile(pgnFilePath)
	defer parser.Finish()

	var outFile *os.File
	if _, err := os.Stat(outFilePath); os.IsNotExist(err) {
		file, err := os.Create(outFilePath)
		if err != nil {
			panic(err)
		}
		outFile = file
	} else {
		file, err := os.OpenFile(outFilePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		outFile = file
	}

	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	_, err := writer.WriteString("fen,outcome\n")

	if err != nil {
		panic(err)
	}

	fens := []string{}
	sd := engine.SearchData{}
	posCopy := engine.Position{}

	sd.Timer.CalculateSearchTime(engine.InfiniteTimeFormat, 0, 0, 0, 0)
	log.Printf("Extracting FENs from %s", pgnFilePath)

	numGames := 0
	for game := parser.NextGame(); game != nil; game = parser.NextGame() {
		numGames++

		if numGames % ReportEveryNGames == 0 {
			log.Printf("%d games scanned\n", numGames)
		} 

		sd.Pos.LoadFEN(game.StartFen)
		gamePly := len(game.Moves)

		if game.Result == NoResult {
			continue
		}

		result := "0.5"
		if game.Result == WhiteWon {
			result = "1.0"
		} else if game.Result == BlackWon {
			result = "0.0"
		}

		fensFromGame := []string{}

		for ply, move := range game.Moves {
			sd.Pos.DoMove(move)

			if ply < 10 || ply > 200 || gamePly-ply <= 10 {
				continue
			}
			
			if sd.Pos.IsSideInCheck(sd.Pos.Side) {
				continue
			}

			score := engine.Qsearch(&sd, -engine.InfinityCPValue, engine.InfinityCPValue, 0)

			if utils.Abs(score) > scoreBoundCP {
				continue
			}

			fen := applyPVToGetFEN(&sd, &posCopy)
			fields := strings.Fields(fen)
			fensFromGame = append(
				fensFromGame, fmt.Sprintf("%s %s - - 0 1, %s\n", fields[0], fields[1], result),
			)
		}

		sampleSize := utils.Min(sampleSizePerGame, uint16(len(fensFromGame)))
		for i := uint16(0); i < sampleSize; i++ {
			fens = append(fens, fensFromGame[rand.Intn(len(fensFromGame))])
		}
	}

	log.Printf("%d fens extracted", len(fens))
	log.Println("Checking for duplicate FENs and shuffling")

	uniqueFENs := []string{}
	seen := make(map[string]bool)
	duplicates := 0

	for _, fen := range fens {
		if seenBefore := seen[fen]; !seenBefore {
			seen[fen] = true
			uniqueFENs = append(uniqueFENs, fen)
		} else {
			duplicates++
		}
	}

	rand.Shuffle(
		len(uniqueFENs), 
		func(i, j int) { uniqueFENs[i], uniqueFENs[j] = uniqueFENs[j], uniqueFENs[i] },
	)


	log.Printf("%d duplicates ignored", duplicates)
	log.Printf("Writing %d FENs to %s", len(uniqueFENs), outFilePath)

	for _, fen := range uniqueFENs {
		_, err := writer.WriteString(fen)
		if err != nil {
			panic(err)
		}
	}

	writer.Flush()
	log.Println("FEN extraction completed successfully")
	log.Printf("%d total games scanned\n", numGames)
}

func applyPVToGetFEN(sd *engine.SearchData, posCopy *engine.Position) string {
	pvLine := sd.GetCurrPV()
	engine.CopyPos(&sd.Pos, posCopy)

	for i := uint8(0); i < pvLine.Cnt; i++ {
		sd.Pos.DoMove(pvLine.Moves[i])
	}

	fen := sd.Pos.GenFEN()
	engine.CopyPos(posCopy, &sd.Pos)
	return fen
} 