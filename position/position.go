package position

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

/*
What functionality does a position object need?
	* Load FEN strings
	* Make moves (unmake moves?)
	* Output FEN strings
	* Print position
	* Given a square can you tell me what piece type and color
	  is sitting there? Can this be done only using bitboards,
	  avoiding using a mailbox representation of the board too? [x]

What needs to be tracked?
	* What pieces are sitting where? Will be done using bitboards,
	  2 for color, 6 for piece type
	* Active color
	* Castling rights
	* En passant square
	* Half-move clock
	* Full-move clock
	* Zobrist hash (eventually, can hasing be done in a a different way if
	  copy/make paradigm is used?)
*/

const (
	PAWN    = 0
	KNIGHT  = 1
	BISHOP  = 2
	ROOK    = 3
	QUEEN   = 4
	KING    = 5
	NO_TYPE = 6

	WHITE    = 0
	BLACK    = 1
	NO_COLOR = 2

	WhiteKingsideRight  uint8 = 0x8
	WhiteQueensideRight uint8 = 0x4
	BlackKingsideRight  uint8 = 0x2
	BlackQueensideRight uint8 = 0x1

	FENStartPosition = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 0"
)

type Position struct {
	Pieces   [6]uint64
	Colors   [2]uint64
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

func (pos *Position) LoadFEN(fen string) {
	pos.Pieces = [6]uint64{}
	pos.Colors = [2]uint64{}

	fields := strings.Fields(fen)
	pieces := fields[0]
	side := fields[1]
	castling := fields[2]
	ep := fields[3]
	halfMove := fields[4]

	for index, sq := 0, A8; index < len(pieces); index++ {
		// rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR
		char := pieces[index]
		switch char {
		case 'p': pos.putPiece(PAWN, BLACK, sq); sq++
		case 'n': pos.putPiece(KNIGHT, BLACK, sq); sq++
		case 'b': pos.putPiece(BISHOP, BLACK, sq); sq++
		case 'r': pos.putPiece(ROOK, BLACK, sq); sq++
		case 'q': pos.putPiece(QUEEN, BLACK, sq); sq++
		case 'k': pos.putPiece(KING, BLACK, sq); sq++
		case 'P': pos.putPiece(PAWN, WHITE, sq); sq++
		case 'N': pos.putPiece(KNIGHT, WHITE, sq); sq++
		case 'B': pos.putPiece(BISHOP, WHITE, sq); sq++
		case 'R': pos.putPiece(ROOK, WHITE, sq); sq++
		case 'Q': pos.putPiece(QUEEN, WHITE, sq); sq++
		case 'K': pos.putPiece(KING, WHITE, sq); sq++
		case '/':
			sq -= 16
		case '1', '2', '3', '4', '5', '6', '7', '8':
			sq += pieces[index] - '0'
		}
	}

	pos.Side = WHITE
	if side == "b" {
		pos.Side = BLACK
	}

	pos.EPSq = NoSq
	if ep != "-" {
		pos.EPSq = coordToSq(ep)
	}

	halfMoveCounter, _ := strconv.Atoi(halfMove)
	pos.HalfMove = uint8(halfMoveCounter)

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
}

func (pos Position) String() (boardStr string) {
	boardStr += "\n"

	for i := 56; i >= 0; i -= 8 {
		boardStr += fmt.Sprintf("%d | ", i/8+1)
		for j := i; j < i+8; j++ {
			pieceType, pieceColor := pos.getPieceOnSq(uint8(j))
			var pieceChar rune

			switch pieceType {
			case PAWN: pieceChar = 'p'
			case KNIGHT: pieceChar = 'n'
			case BISHOP: pieceChar = 'b'
			case ROOK: pieceChar = 'r'
			case QUEEN: pieceChar = 'q'
			case KING: pieceChar = 'k'
			case NO_TYPE: pieceChar = '.'
			}

			if pieceColor == WHITE {
				pieceChar = unicode.ToUpper(pieceChar)
			}

			boardStr += fmt.Sprintf("%c ", pieceChar)
		}
		boardStr += "\n"
	}

	boardStr += "   ----------------"
	boardStr += "\n    a b c d e f g h"

	boardStr += "\n\n"
	if pos.Side == WHITE {
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
		boardStr += sqToCoord(pos.EPSq)
	}

	boardStr += fmt.Sprintf("\nhalf-move clock: %d\n", pos.HalfMove)
	return boardStr
}

func (pos *Position) putPiece(pieceType, pieceColor, sq uint8) {
	pos.Pieces[pieceType] = setBit(pos.Pieces[pieceType], sq)
	pos.Colors[pieceColor] = setBit(pos.Colors[pieceColor], sq)
}

func (pos Position) getPieceOnSq(sq uint8) (pieceType, pieceColor uint8) {
	empty := ^(pos.Colors[WHITE] | pos.Colors[BLACK])

	pawnType := ((pos.Pieces[PAWN] & SquareBB[sq]) >> sq) * PAWN
	knightType := ((pos.Pieces[KNIGHT] & SquareBB[sq]) >> sq) * KNIGHT
	bishopType := ((pos.Pieces[BISHOP] & SquareBB[sq]) >> sq) * BISHOP
	rookType := ((pos.Pieces[ROOK] & SquareBB[sq]) >> sq) * ROOK
	queenType := ((pos.Pieces[QUEEN] & SquareBB[sq]) >> sq) * QUEEN
	kingType := ((pos.Pieces[KING] & SquareBB[sq]) >> sq) * KING
	noType := ((empty & SquareBB[sq]) >> sq) * NO_TYPE

	whiteColor := ((pos.Colors[WHITE] & SquareBB[sq]) >> sq) * WHITE
	blackColor := ((pos.Colors[BLACK] & SquareBB[sq]) >> sq) * BLACK
	noColor := ((empty & SquareBB[sq]) >> sq) * NO_COLOR

	return uint8(pawnType + knightType + bishopType + rookType + queenType + kingType + noType),
		uint8(whiteColor + blackColor + noColor)
}