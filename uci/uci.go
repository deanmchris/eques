package uci

import (
	"bufio"
	"bullet/engine"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	EngineName = "Bullet 1.0.0"
	EngineAuthor = "Christian Dean"
)

type GameData struct {
	numOfMoves    uint16
	initValuesSet bool
}

func (g *GameData) Reset() {
	g.numOfMoves = 0
	g.initValuesSet = false
}

type TokensQueue struct {
	tokens []string
}

func (tq *TokensQueue) Add(token string) {
	tq.tokens = append(tq.tokens, token)
}

func (tq *TokensQueue) Pop() string {
	token := tq.tokens[0] 
	tq.tokens = tq.tokens[1:]
	return token
}

func (tq *TokensQueue) Size() int {
	return len(tq.tokens)
}

func UCICommandReponse() {
	fmt.Printf("id name %v\n", EngineName)
	fmt.Printf("id author %v\n", EngineAuthor)
	fmt.Println("uciok")
}

func isReadyCommandReponse() {
	fmt.Println("readyok")
}

func UCINewGameCommandReponse(sd *engine.SearchData, g *GameData) {
	sd.Reset()
	g.Reset()
}

func positionCommandReponse(sd *engine.SearchData, g *GameData, tokens *TokensQueue) {
	token := tokens.Pop()
	if token == "fen" {
		fenStringBuilder := strings.Builder{}
		for i := 0; i < 6; i++ {
			fenStringBuilder.WriteString(tokens.Pop())
			fenStringBuilder.WriteString(" ")
		}
		fenString := strings.TrimSpace(fenStringBuilder.String())
		sd.Pos.LoadFEN(fenString)
	} else if token == "startpos" {
		sd.Pos.LoadFEN(engine.FENStartPosition)
	}

	sd.ClearPosHistory()
	sd.AddCurrPosToHistory()
	g.numOfMoves = 0

	if tokens.Size() > 0 && tokens.Pop() == "moves" {
		for tokens.Size() > 0 {
			moveToken := tokens.Pop()
			move := parseUCIMove(sd, moveToken)
			sd.Pos.DoMove(move)
			sd.AddCurrPosToHistory()
			g.numOfMoves++
		}
	}
	
	g.numOfMoves /= 2
}

func goCommandReponse(sd *engine.SearchData, g *GameData, tokens *TokensQueue) {
	prefix := "b"
	if sd.Pos.Side == engine.White {
		prefix = "w"
	}

	timeFormat := engine.NoFormat
	timeLeft := int64(0)
	timeInc := int64(0)
	movesToGo := int64(0)

	for tokens.Size() > 0 {
		token := tokens.Pop()
		switch token {
		case prefix+"inc":
			timeInc = int64(parseInt(tokens.Pop()))
		case prefix+"time":
			timeLeft = int64(parseInt(tokens.Pop()))
			if timeFormat == engine.NoFormat {
				timeFormat = engine.SuddenDeathTimeFormat
			}
		case "movestogo":
			movesToGo = int64(parseInt(tokens.Pop()))
			timeFormat = engine.MovesToGoTimingFormat
		case "infinite":
			timeFormat = engine.InfiniteTimeFormat
		}
	}

	if !g.initValuesSet {
		sd.Timer.SetInitValues(timeFormat, movesToGo)
		g.initValuesSet = true
	}

	sd.Timer.CalculateSearchTime(timeFormat, movesToGo, timeLeft, timeInc, g.numOfMoves)
	bestMove := engine.Search(sd)
	fmt.Printf("bestmove %v\n", bestMove)
}

func stopCommandReponse(sd *engine.SearchData) {
	sd.Timer.Stopped = true
}

func parseUCIMove(sd *engine.SearchData, move string) engine.Move {
	fromSq := engine.CoordToSq(move[0:2])
	toSq := engine.CoordToSq(move[2:4])
	promoFlag := move[4:]

	pieceType := sd.Pos.GetPieceTypeOnSq(fromSq)
	attackedType := sd.Pos.GetPieceTypeOnSq(toSq)
	moveType := uint8(0)

	if promoFlag == "n" && attackedType != engine.NoType {
		moveType = engine.PromoAttkN
	} else if promoFlag == "b" && attackedType != engine.NoType {
		moveType = engine.PromoAttkB
	} else if promoFlag == "r" && attackedType != engine.NoType {
		moveType = engine.PromoAttkR
	} else if promoFlag == "q" && attackedType != engine.NoType {
		moveType = engine.PromoAttkQ
	} else if promoFlag == "n" {
		moveType = engine.PromoN
	} else if promoFlag == "b" {
		moveType = engine.PromoB
	} else if promoFlag == "r" {
		moveType = engine.PromoR
	} else if promoFlag == "q" {
		moveType = engine.PromoQ
	} else if move == "e1g1" && pieceType == engine.King {
		moveType = engine.WhiteCastleK
	} else if move == "e1c1" && pieceType == engine.King {
		moveType = engine.WhiteCastleQ
	} else if move == "e8g8" && pieceType == engine.King {
		moveType = engine.BlackCastleK
	} else if move == "e8c8" && pieceType == engine.King {
		moveType = engine.BlackCastleQ
	} else if pieceType == engine.Pawn && toSq == sd.Pos.EPSq && sd.Pos.Side == engine.White {
		moveType = engine.WhiteAttackEP
	} else if pieceType == engine.Pawn && toSq == sd.Pos.EPSq && sd.Pos.Side == engine.Black {
		moveType = engine.BlackAttackEP
	} else if attackedType != engine.NoType {
		moveType = engine.Attack
	} else {
		moveType = engine.Quiet
	}

	return engine.NewMove(fromSq, toSq, pieceType, moveType)
}

func parseInt(intAsStr string) int {
	val, err := strconv.Atoi(intAsStr)
	if err != nil {
		panic(fmt.Sprintf("error converting int %v to integer datatype\n", intAsStr))
	}
	return val
}

func parseTokens(input string) (tokens TokensQueue) {
	for _, token := range strings.Fields(input) {
		tokens.Add(sanatizeString(token, "\r\n "))
	}
	return tokens
}

func sanatizeString(input,  removedChars string) string {
	for _, char := range removedChars {
		input = strings.Replace(input, "\n", string(char), -1)
	}
	return input
}

func StartUCIProtocolInterface() {
	reader := bufio.NewReader(os.Stdin)
	searchData := engine.SearchData{}
	gameData := GameData{}

	UCICommandReponse()
	searchData.Pos.LoadFEN(engine.FENStartPosition)
	searchData.AddCurrPosToHistory()

	for {
		command, _ := reader.ReadString('\n')
		command = sanatizeString(command, "\n\r")
		tokens := parseTokens(command)

		switch tokens.Pop() {
		case "uci":
			UCICommandReponse()
		case "isready":
			isReadyCommandReponse()
		case "ucinewgame":
			UCINewGameCommandReponse(&searchData, &gameData)
		case "position":
			positionCommandReponse(&searchData, &gameData, &tokens)
		case "go":
			go goCommandReponse(&searchData, &gameData, &tokens)
		case "stop":
			stopCommandReponse(&searchData)
		case "quit":
			return
		}
	}
}
