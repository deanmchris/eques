package engine

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	Pawn    = 0
	Knight  = 1
	Bishop  = 2
	Rook    = 3
	Queen   = 4
	King    = 5
	NoType = 6

	White    = 0
	Black    = 1
	NoColor = 2

	WhiteKingsideRight  uint8 = 0x8
	WhiteQueensideRight uint8 = 0x4
	BlackKingsideRight  uint8 = 0x2
	BlackQueensideRight uint8 = 0x1

	FENStartPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 0"

	PositionStackSize = 80
)

var Spoilers = [64]uint8{
	0xb, 0xf, 0xf, 0xf, 0x3, 0xf, 0xf, 0x7,
	0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf,
	0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf,
	0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf,
	0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf,
	0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf,
	0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf,
	0xe, 0xf, 0xf, 0xf, 0xc, 0xf, 0xf, 0xd,
}

type Position struct {
	Pieces   [6]uint64
	Colors   [2]uint64
	Hash     uint64
	Scores   [2]int16
	Side,
	Castling,
	EPSq,
	HalfMove uint8
}

func NewPosition(fen string) Position {
	pos := Position{}
	pos.LoadFEN(fen)
	return pos
}

func CopyPos(oldPos, newPos *Position) {
	newPos.Pieces = oldPos.Pieces
	newPos.Colors = oldPos.Colors
	newPos.Hash = oldPos.Hash
	newPos.Scores = oldPos.Scores
	newPos.Side = oldPos.Side
	newPos.Castling = oldPos.Castling
	newPos.EPSq = oldPos.EPSq
	newPos.HalfMove = oldPos.HalfMove
}

func (pos *Position) LoadFEN(fen string) {
	pos.Pieces = [6]uint64{}
	pos.Colors = [2]uint64{}
	pos.Scores = [2]int16{}

	fields := strings.Fields(fen)
	pieces := fields[0]
	side := fields[1]
	castling := fields[2]
	ep := fields[3]
	halfMove := fields[4]

	for index, sq := 0, A8; index < len(pieces); index++ {
		char := pieces[index]
		switch char {
		case 'p': pos.putPiece(Pawn, Black, sq); sq++
		case 'n': pos.putPiece(Knight, Black, sq); sq++
		case 'b': pos.putPiece(Bishop, Black, sq); sq++
		case 'r': pos.putPiece(Rook, Black, sq); sq++
		case 'q': pos.putPiece(Queen, Black, sq); sq++
		case 'k': pos.putPiece(King, Black, sq); sq++
		case 'P': pos.putPiece(Pawn, White, sq); sq++
		case 'N': pos.putPiece(Knight, White, sq); sq++
		case 'B': pos.putPiece(Bishop, White, sq); sq++
		case 'R': pos.putPiece(Rook, White, sq); sq++
		case 'Q': pos.putPiece(Queen, White, sq); sq++
		case 'K': pos.putPiece(King, White, sq); sq++
		case '/':
			sq -= 16
		case '1', '2', '3', '4', '5', '6', '7', '8':
			sq += pieces[index] - '0'
		}
	}

	pos.Side = White
	if side == "b" {
		pos.Side = Black
	}

	pos.EPSq = NoSq
	if ep != "-" {
		pos.EPSq = CoordToSq(ep)
	}

	halfMoveCounter, _ := strconv.Atoi(halfMove)
	pos.HalfMove = uint8(halfMoveCounter)

	pos.Castling = 0
	for _, char := range castling {
		switch char {
		case 'K':
			pos.Castling |= WhiteKingsideRight
		case 'Q':
			pos.Castling |= WhiteQueensideRight
		case 'k':
			pos.Castling |= BlackKingsideRight
		case 'q':
			pos.Castling |= BlackQueensideRight
		}
	}

	pos.Hash = GenHash(pos)
}

func (pos Position) String() (boardStr string) {
	boardStr += "\n"

	for i := 56; i >= 0; i -= 8 {
		boardStr += fmt.Sprintf("%d | ", i/8+1)
		for j := i; j < i+8; j++ {
			pieceType := pos.GetPieceTypeOnSq(uint8(j))
		    pieceColor := pos.GetPieceColorOnSq(uint8(j))
			boardStr += fmt.Sprintf("%c ", getPieceCharFromType(pieceType, pieceColor))
		}
		boardStr += "\n"
	}

	boardStr += "   ----------------"
	boardStr += "\n    a b c d e f g h"

	boardStr += "\n\n"
	if pos.Side == White {
		boardStr += "turn: white\n"
	} else {
		boardStr += "turn: black\n"
	}

	boardStr += "castling rights: "
	if pos.Castling&WhiteKingsideRight != 0 {
		boardStr += "K"
	}
	if pos.Castling&WhiteQueensideRight != 0 {
		boardStr += "Q"
	}
	if pos.Castling&BlackKingsideRight != 0 {
		boardStr += "k"
	}
	if pos.Castling&BlackQueensideRight != 0 {
		boardStr += "q"
	}

	boardStr += "\nen passant: "
	if pos.EPSq == NoSq {
		boardStr += "none"
	} else {
		boardStr += SqToCoord(pos.EPSq)
	}

	boardStr += fmt.Sprintf("\nhalf-move clock: %d", pos.HalfMove)
	boardStr += fmt.Sprintf("\nzobrist hash: 0x%x\n", pos.Hash)
	return boardStr
}

func (pos *Position) GenFEN() string {
	positionStr := strings.Builder{}
	for rankStartPos := 56; rankStartPos >= 0; rankStartPos -= 8 {
		emptySquares := 0
		for sq := rankStartPos; sq < rankStartPos+8; sq++ {
			pieceType := pos.GetPieceTypeOnSq(uint8(sq))
			pieceColor := pos.GetPieceColorOnSq(uint8(sq))

			if pieceType == NoType {
				emptySquares++
			} else {
				if emptySquares > 0 {
					positionStr.WriteString(strconv.Itoa(emptySquares))
					emptySquares = 0
				}

				positionStr.WriteRune(getPieceCharFromType(pieceType, pieceColor))
			}
		}
		if emptySquares > 0 {
			positionStr.WriteString(strconv.Itoa(emptySquares))
			emptySquares = 0
		}
		positionStr.WriteString("/")
	}

	sideToMove := ""
	castlingRights := ""
	epSquare := ""

	if pos.Side == White {
		sideToMove = "w"
	} else {
		sideToMove = "b"
	}

	if pos.Castling&WhiteKingsideRight != 0 {
		castlingRights += "K"
	}
	if pos.Castling&WhiteQueensideRight != 0 {
		castlingRights += "Q"
	}
	if pos.Castling&BlackKingsideRight != 0 {
		castlingRights += "k"
	}
	if pos.Castling&BlackQueensideRight != 0 {
		castlingRights += "q"
	}

	if castlingRights == "" {
		castlingRights = "-"
	}

	if pos.EPSq == NoSq {
		epSquare = "-"
	} else {
		epSquare = SqToCoord(pos.EPSq)
	}

	return fmt.Sprintf(
		"%s %s %s %s %d %d",
		strings.TrimSuffix(positionStr.String(), "/"),
		sideToMove, castlingRights, epSquare,
		pos.HalfMove, 0,
	)
}

func (pos *Position) IsSideInCheck(side uint8) bool {
	return pos.SqIsAttacked(side, GetLSBpos(pos.Pieces[King] & pos.Colors[side]))
}

func (pos *Position) SqIsAttacked(usColor, sq uint8) bool {
	enemyBB := pos.Colors[usColor^1]
	usBB := pos.Colors[usColor]

	enemyKnights := pos.Pieces[Knight] & enemyBB
	enemyKing := pos.Pieces[King] & enemyBB
	enemyPawns := pos.Pieces[Pawn] & enemyBB
	
	if KnightMoves[sq]&enemyKnights != 0 {
		return true
	}
	if KingMoves[sq]&enemyKing != 0 {
		return true
	}
	if PawnAttacks[usColor][sq]&enemyPawns != 0 {
		return true
	}

	enemyBishops := pos.Pieces[Bishop] & enemyBB
	enemyRooks := pos.Pieces[Rook] & enemyBB
	enemyQueens := pos.Pieces[Queen] & enemyBB

	intercardinalRays := LookupBishopMoves(sq, enemyBB|usBB)
	cardinalRaysRays := LookupRookMoves(sq, enemyBB|usBB)

	if intercardinalRays&(enemyBishops|enemyQueens) != 0 {
		return true
	}
	if cardinalRaysRays&(enemyRooks|enemyQueens) != 0 {
		return true
	}
	
	return false
}

func (pos *Position) DoMove(move Move) {
	toSq := move.ToSq()
	fromSq := move.FromSq()
	pieceType := move.FromType()

	pos.Hash ^= EPSqZobristValues[pos.EPSq]
	pos.Hash ^= CastlingZobristValues[pos.Castling]
	pos.Hash ^= SideZobristValues[pos.Side]

	pos.removePiece(pieceType, pos.Side, fromSq)

	pos.HalfMove++
	pos.EPSq = NoSq

	switch move.Type() {
	case Quiet: pos.putPiece(pieceType, pos.Side, toSq)
	case Attack: pos.doAttack(pieceType, toSq)
	case WhiteAttackEP: pos.doEPAttack(toSq, toSq-8)
	case BlackAttackEP: pos.doEPAttack(toSq, toSq+8)
	case PromoQ: pos.putPiece(Queen, pos.Side, toSq)
	case PromoR: pos.putPiece(Rook, pos.Side, toSq)
	case PromoB: pos.putPiece(Bishop, pos.Side, toSq)
	case PromoN: pos.putPiece(Knight, pos.Side, toSq)
	case PromoAttkQ: pos.doAttack(Queen, toSq)
	case PromoAttkR: pos.doAttack(Rook, toSq)
	case PromoAttkB: pos.doAttack(Bishop, toSq)
	case PromoAttkN: pos.doAttack(Knight, toSq)
	case WhiteCastleK: pos.doCastle(G1, H1, F1)
	case WhiteCastleQ: pos.doCastle(C1, A1, D1)
	case BlackCastleK: pos.doCastle(G8, H8, F8)
	case BlackCastleQ: pos.doCastle(C8, A8, D8)
	}

	if pieceType == Pawn {
		pos.HalfMove = 0
		if pos.Side == White && toSq-fromSq == 16 {
			pos.EPSq = toSq-8
		} else if pos.Side == Black && fromSq-toSq == 16 {
			pos.EPSq = toSq+8
		}
	}
	
	pos.Castling &= Spoilers[fromSq] & Spoilers[toSq]
	pos.Side ^= 1

	pos.Hash ^= EPSqZobristValues[pos.EPSq]
	pos.Hash ^= CastlingZobristValues[pos.Castling]
	pos.Hash ^= SideZobristValues[pos.Side]
}

func (pos *Position) doEPAttack(toSq, capturedPawnSq uint8) {
	pos.removePiece(Pawn, pos.Side^1, capturedPawnSq)
	pos.putPiece(Pawn, pos.Side, toSq)
	pos.HalfMove = 0
}

func (pos *Position) doAttack(typeOnToSq, toSq uint8) {
	attackedType := pos.GetPieceTypeOnSq(toSq)
	pos.removePiece(attackedType, pos.Side^1, toSq)
	pos.putPiece(typeOnToSq, pos.Side, toSq)
	pos.HalfMove = 0
}

func (pos *Position) doCastle(kingToSq, rookFromSq, rookToSq uint8) {
	pos.removePiece(Rook, pos.Side, rookFromSq)
	pos.putPiece(King, pos.Side, kingToSq)
	pos.putPiece(Rook, pos.Side, rookToSq)
}


func (pos *Position) putPiece(pieceType, pieceColor, sq uint8) {
	pos.Pieces[pieceType] = SetBit(pos.Pieces[pieceType], sq)
	pos.Colors[pieceColor] = SetBit(pos.Colors[pieceColor], sq)
	pos.Hash ^= PieceZobristValues[pieceColor][pieceType][sq]
	pos.Scores[pieceColor] += PieceSquareTable[pieceType][FlipSq[pieceColor][sq]]
}

func (pos *Position) removePiece(pieceType, pieceColor, sq uint8) {
	pos.Pieces[pieceType] = UnsetBit(pos.Pieces[pieceType], sq)
	pos.Colors[pieceColor] = UnsetBit(pos.Colors[pieceColor], sq)
	pos.Hash ^= PieceZobristValues[pieceColor][pieceType][sq]
	pos.Scores[pieceColor] -= PieceSquareTable[pieceType][FlipSq[pieceColor][sq]]
}

func (pos *Position) GetPieceTypeOnSq(sq uint8) uint8 {
	sqBB := uint64(1) << sq
	if pos.Pieces[Pawn] & sqBB != 0 {
		return Pawn
	} else if pos.Pieces[Knight] & sqBB != 0 {
		return Knight
	} else if pos.Pieces[Bishop] & sqBB != 0 {
		return Bishop
	} else if pos.Pieces[Rook] & sqBB != 0 {
		return Rook
	} else if pos.Pieces[Queen] & sqBB != 0 {
		return Queen
	} else if pos.Pieces[King] & sqBB != 0 {
		return King
	} 
	return NoType
}

func (pos *Position) GetPieceColorOnSq(sq uint8) uint8 {
	sqBB := uint64(1) << sq
	if pos.Colors[White] & sqBB != 0 {
		return White
	} else if pos.Colors[Black] & sqBB != 0 {
		return Black
	} 
	return NoColor
}

func getPieceCharFromType(pieceType, pieceColor uint8) rune {
	var pieceChar rune
	switch pieceType {
	case Pawn: pieceChar = 'p'
	case Knight: pieceChar = 'n'
	case Bishop: pieceChar = 'b'
	case Rook: pieceChar = 'r'
	case Queen: pieceChar = 'q'
	case King: pieceChar = 'k'
	case NoType: pieceChar = '.'
	}

	if pieceColor == White {
		pieceChar = unicode.ToUpper(pieceChar)
	}

	return pieceChar
}