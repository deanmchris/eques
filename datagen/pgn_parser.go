package datagen

import (
	"bufio"
	"bullet/engine"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	WhiteWon uint8 = 0
	BlackWon uint8 = 1
	Drawn    uint8 = 3
	NoResult uint8 = 4

	EventTagPattern  = "\\[Event .*\\]"
	FenTagPattern    = "\\[FEN (.*)\\]"
	ResultTagPattern = "\\[Result (.*)\\]"
	TagPattern       = "\\[.*\\]"
	MovePattern      = "(?P<pawn_push>[a-h][2-7])|" +
	                   "(?P<pawn_cap_promo>[a-h]x[a-h][1-8]=[QRBN])|" +
	                   "(?P<pawn_cap>[a-h]x[a-h][1-8])|" +
					   "(?P<pawn_promo>[a-h][18]=[QRBN])|" +
					   "(?P<quiet_full>[QRBN][a-h][1-8][a-h][1-8])|" +
					   "(?P<cap_full>[QRBN][a-h][1-8]x[a-h][1-8])|" +
					   "(?P<quiet>[QRBNK][a-h][1-8])|" +
					   "(?P<quiet_file>[QRBN][a-h][a-h][1-8])|" +
					   "(?P<quiet_rank>[QRBN][1-8][a-h][1-8])|" +
					   "(?P<cap>[QRBNK]x[a-h][1-8])|" +
					   "(?P<cap_file>[QRBN][a-h]x[a-h][1-8])|" +
					   "(?P<cap_rank>[QRBN][1-8]x[a-h][1-8])|" +
					   "(?P<castle_qs>O-O-O)|" +
					   "(?P<castle_ks>O-O)"

	ArgNotUsed uint8 = 8
)

const (
	PawnPush = iota + 1 // pawn push
	PawnCapPromo        // pawn promotion and capture
	PawnCap             // pawn capture
	PawnPromo           // pawn promotion
	QuietFull           // non-pawn quiet move disambiguated with departing file and rank
	CapFull             // non-pawn capture disambiguated with departing file and rank
	Quiet               // non-pawn quiet move
	QuietFile           // non-pawn quiet move disambiguated with departing file
	QuietRank           // non-pawn quiet move disambiguated with departing rank
	Cap                 // non-pawn capture
	CapFile             // non-pawn capture disambiguated with departing file
	CapRank             // non-pawn capture disambiguated with departing rank
	CastleQS            // castle queen-side
	CastleKS            // castle king-side
)

type Game struct {
	Moves    []engine.Move
	StartFen string
	Result   uint8
}


type PGNParser struct {
	pos            engine.Position

	// It's helpful to have a position to copy the current position into
	// when we need to make a move and then undo it.
	posCopy        engine.Position

	pgnFile        *os.File
	scanner        *bufio.Scanner
	eventTagRegex  *regexp.Regexp
	fenTagRegex    *regexp.Regexp
	resultTagRegex *regexp.Regexp
	tagRegex       *regexp.Regexp
	moveRegex      *regexp.Regexp
}

func (parser *PGNParser) LoadPGNFile(path string) {
	file, err := os.Open(path)
	parser.pgnFile = file
	parser.scanner = bufio.NewScanner(file)

	if err != nil {
		panic(err)
	}

	parser.eventTagRegex  = regexp.MustCompile(EventTagPattern)
	parser.fenTagRegex    = regexp.MustCompile(FenTagPattern)
	parser.resultTagRegex = regexp.MustCompile(ResultTagPattern)
	parser.tagRegex       = regexp.MustCompile(TagPattern)
	parser.moveRegex      = regexp.MustCompile(MovePattern)

	for parser.scanner.Scan() {
    	line := parser.scanner.Text()
		if parser.eventTagRegex.MatchString(line) {
			break
		}
    }
}

func (parser *PGNParser) NextGame() *Game {
	game := Game{StartFen: engine.FENStartPosition, Result: NoResult}
	pgnMoves := strings.Builder{}
	done := true

    for parser.scanner.Scan() {
		line := parser.scanner.Text()
		done = false

		if parser.eventTagRegex.MatchString(line) {
			break
		}
		
		line = strings.TrimSpace(line)

		if line == "\n" {
			continue
		}

		fenTagMatch := parser.fenTagRegex.FindStringSubmatch(line)
		if len(fenTagMatch) > 0 {
			fen := fenTagMatch[1]
			fen = strings.TrimPrefix(fen, "\"")
			fen = strings.TrimSuffix(fen, "\"")
			game.StartFen = fen
			continue
		}

		resultTagMatch := parser.resultTagRegex.FindStringSubmatch(line)
		if len(resultTagMatch) > 0 {
			result := resultTagMatch[1]
			if result == "\"1-0\"" {
				game.Result = WhiteWon
			} else if result == "\"0-1\"" {
				game.Result = BlackWon
			} else if result == "\"1/2-1/2\"" {
				game.Result = Drawn
			}
		}

		if parser.tagRegex.MatchString(line) {
			continue
		}

        pgnMoves.WriteString(line)
    }

	parser.pos.LoadFEN(game.StartFen)
	matches := parser.moveRegex.FindAllStringSubmatch(pgnMoves.String(), -1)

	for _, match := range matches {
		subexpIndex := uint8(0)
		for subexpIndex = PawnPush; subexpIndex <= CastleQS; subexpIndex++ {
			if match[subexpIndex] != "" {
				break
			}
		}

		move := parseSANToMove(&parser.pos, &parser.posCopy, match[0], subexpIndex)
		parser.pos.DoMove(move)
		game.Moves = append(game.Moves, move)
	}

	if done {
		return nil
	}

	return &game
}

func (parser *PGNParser) Finish() {
	parser.pgnFile.Close()
}

func parseSANToMove(pos, posCopy *engine.Position, move string, moveType uint8) engine.Move {
	switch moveType {
	case PawnPush:
		toSq := engine.CoordToSq(move)
		fromSq := getPawnPushFromSq(pos.Pieces[engine.Pawn] & pos.Colors[pos.Side], pos.Side, toSq)
		return engine.NewMove(fromSq, toSq, engine.Pawn, engine.Quiet)
	case PawnCapPromo:
		toSq := engine.CoordToSq(move[2:4])
		fromSq := getPawnAttackFromSq(pos.Side, move[0], move[3])
		return engine.NewMove(
			fromSq, toSq, engine.Pawn, 
			getPromoMoveType(move[5])+engine.DeltaToGenerateAttackPromotions,
		)
	case PawnCap:
		toSq := engine.CoordToSq(move[2:4])
		fromSq := getPawnAttackFromSq(pos.Side, move[0], move[3])
		if toSq == pos.EPSq && pos.Side == engine.White {
			return engine.NewMove(fromSq, toSq, engine.Pawn, engine.WhiteAttackEP)
		} else if toSq == pos.EPSq && pos.Side == engine.Black {
			return engine.NewMove(fromSq, toSq, engine.Pawn, engine.BlackAttackEP)
		}
		return engine.NewMove(fromSq, toSq, engine.Pawn, engine.Attack)
	case PawnPromo:
		toSq := engine.CoordToSq(move[0:2])
		fromSq := getPawnPushFromSq(pos.Pieces[engine.Pawn] & pos.Colors[pos.Side], pos.Side, toSq)
		return engine.NewMove(
			fromSq, toSq, engine.Pawn, 
			getPromoMoveType(move[3])+engine.DeltaToGenerateQuietPromotions,
		)
	case Quiet:
		return genGeneralPieceMoveFromIncompleteInfo(
			pos, posCopy, engine.CoordToSq(move[1:3]), pieceCharToType(move[0]),
			ArgNotUsed, ArgNotUsed, engine.Quiet,
		)
	case QuietFile:
		return genGeneralPieceMoveFromIncompleteInfo(
			pos, posCopy, engine.CoordToSq(move[2:4]), pieceCharToType(move[0]),
			ArgNotUsed, fileCharToInt(move[1]), engine.Quiet,
		)
	case QuietRank:
		return genGeneralPieceMoveFromIncompleteInfo(
			pos, posCopy, engine.CoordToSq(move[2:4]), pieceCharToType(move[0]),
			rankCharToInt(move[1]), ArgNotUsed, engine.Quiet,
		)
	case QuietFull:
		toSq := engine.CoordToSq(move[3:5])
		fromSq := engine.CoordToSq(move[1:3])
		pieceType := pieceCharToType(move[0])
		return engine.NewMove(fromSq, toSq, pieceType, engine.Quiet)
	case Cap:
		return genGeneralPieceMoveFromIncompleteInfo(
			pos, posCopy, engine.CoordToSq(move[2:4]), pieceCharToType(move[0]),
			ArgNotUsed, ArgNotUsed, engine.Attack,
		)
	case CapFile:
		return genGeneralPieceMoveFromIncompleteInfo(
			pos, posCopy, engine.CoordToSq(move[3:5]), pieceCharToType(move[0]),
			ArgNotUsed, fileCharToInt(move[1]), engine.Attack,
		)
	case CapRank:
		return genGeneralPieceMoveFromIncompleteInfo(
			pos, posCopy, engine.CoordToSq(move[3:5]), pieceCharToType(move[0]),
			rankCharToInt(move[1]), ArgNotUsed, engine.Attack,
		)
	case CapFull:
		toSq := engine.CoordToSq(move[4:6])
		fromSq := engine.CoordToSq(move[1:3])
		pieceType := pieceCharToType(move[0])
		return engine.NewMove(fromSq, toSq, pieceType, engine.Attack)
	case CastleKS:
		if pos.Side == engine.White {
			return engine.NewMove(engine.E1, engine.G1, engine.King, engine.WhiteCastleK)
		}
		return engine.NewMove(engine.E8, engine.G8, engine.King, engine.BlackCastleK)
	case CastleQS:
		if pos.Side == engine.White {
			return engine.NewMove(engine.E1, engine.C1, engine.King, engine.WhiteCastleQ)
		}
		return engine.NewMove(engine.E8, engine.C8, engine.King, engine.BlackCastleQ)
	}

	panic(fmt.Errorf("unknown move type: %d", moveType))
}

func genGeneralPieceMoveFromIncompleteInfo(pos, copyPos *engine.Position, toSq, pieceType, rank, file, moveType uint8) engine.Move {
	movesBB := getMovesFromSq(toSq, pieceType, pos.Colors[engine.White] | pos.Colors[engine.Black])
	var fromSquares []uint8

	if file != ArgNotUsed {
		fromSquares = getFromSqGivenMovesAndFromFile(pos.Pieces[pieceType] & pos.Colors[pos.Side], movesBB, file)
	} else if rank != ArgNotUsed {
		fromSquares = getFromSqGivenMovesAndFromRank(pos.Pieces[pieceType] & pos.Colors[pos.Side], movesBB, rank)
	} else {
		fromSquares = getFromSqGivenMoves(pos.Pieces[pieceType] & pos.Colors[pos.Side], movesBB)
	}

	// Up to this point, the only ambiguity we're not accounting for is the kind that can occur when two
	// pieces can move to the same square, but one is pinned. This means no file or rank will be given
	// in the SAN, since one piece being pinned is enough to disambiguate which the from square of the
	// moving piece. 
	//
	// The easiest way for us to account for this is simply to create each potential move and play it 
	// in the current position. The first one that doesn't leave the king in check must be the correct one.

	for _, fromSq := range fromSquares {
		move := engine.NewMove(fromSq, toSq, pieceType, moveType)
		engine.CopyPos(pos, copyPos)
		pos.DoMove(move)

		if !pos.IsSideInCheck(pos.Side^1) {
			engine.CopyPos(copyPos, pos)
			return move
		}
		engine.CopyPos(copyPos, pos)
	}	

	panic("no valid from square found for given move")
}

func getMovesFromSq(fromSq uint8, pieceType uint8, allBB uint64) uint64 {
	switch pieceType {
	case engine.Knight: return engine.KnightMoves[fromSq]
	case engine.Bishop: return engine.LookupBishopMoves(fromSq, allBB)
	case engine.Rook: return engine.LookupRookMoves(fromSq, allBB)
	case engine.Queen:
		intercardinalMoves := engine.LookupBishopMoves(fromSq, allBB)
		cardinalMoves := engine.LookupRookMoves(fromSq, allBB)
		return intercardinalMoves | cardinalMoves
	case engine.King:
		return engine.KingMoves[fromSq]
	}

	panic(fmt.Errorf("unknown piece type: %d", pieceType))
}

func getFromSqGivenMoves(pieceBB, movesBB uint64) (fromSquares []uint8) {
	sqBB := pieceBB & movesBB
	for sqBB != 0 {
		fromSquares = append(fromSquares, engine.GetLSBpos(sqBB))
		sqBB &= (sqBB - 1)
	}
	return fromSquares
}

func getFromSqGivenMovesAndFromFile(pieceBB, movesBB uint64, file uint8) (fromSquares []uint8) {
	sqBB := pieceBB & movesBB & engine.MaskFile[file]
	for sqBB != 0 {
		fromSquares = append(fromSquares, engine.GetLSBpos(sqBB))
		sqBB &= (sqBB - 1)
	}
	return fromSquares
}

func getFromSqGivenMovesAndFromRank(pieceBB, movesBB uint64, rank uint8) (fromSquares []uint8) {
	sqBB := pieceBB & movesBB & engine.MaskRank[rank]
	for sqBB != 0 {
		fromSquares = append(fromSquares, engine.GetLSBpos(sqBB))
		sqBB &= (sqBB - 1)
	}
	return fromSquares
}

func getPawnPushFromSq(pawnBB uint64, side uint8, toSq uint8) uint8 {
	if side == engine.White {
		if !engine.IsBitset(pawnBB, toSq - 8) {
			return toSq - 16
		}
		return toSq - 8
	}
	if !engine.IsBitset(pawnBB, toSq + 8) {
		return toSq + 16
	}
	return toSq + 8
}

func getPawnAttackFromSq(side uint8, fromFile, toRank byte) uint8 {
	if side == engine.White {
		return engine.CoordToSq(string(fromFile)+string(toRank-1))
	}
	return engine.CoordToSq(string(fromFile)+string(toRank+1))
}

func getPromoMoveType(promoChar byte) uint8 {
	switch promoChar {
	case 'Q': return engine.PromoQ
	case 'R': return engine.PromoR
	case 'B': return engine.PromoB
	case 'N': return engine.PromoN
	}
	
	panic(fmt.Errorf("unknown promotion character: %c", promoChar))
}

func pieceCharToType(pieceChar byte) uint8 {
	switch pieceChar {
	case 'N': return engine.Knight
	case 'B': return engine.Bishop
	case 'R': return engine.Rook
	case 'Q': return engine.Queen
	case 'K': return engine.King
	}

	panic(fmt.Errorf("unknown piece character: %c", pieceChar))
}


func fileCharToInt(fileChar byte) uint8 {
	return uint8(fileChar - 'a')
}

func rankCharToInt(rankChar byte) uint8 {
	return uint8(rankChar-'0') - 1
}