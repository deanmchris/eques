package position

import (
	"bullet/move"
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

			boardStr += fmt.Sprintf("%c ", pieceChar)
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

	boardStr += fmt.Sprintf("\nhalf-move clock: %d\n", pos.HalfMove)
	return boardStr
}

func (pos Position) DoMove(mv move.Move) Position {
	toSq := mv.ToSq()
	fromSq := mv.FromSq()

	pieceType, pieceColor := pos.getPieceOnSq(fromSq)
	pos.removePiece(pieceType, pieceColor, fromSq)

	pos.HalfMove++

	switch mv.Type() {
	case move.Quiet: pos.putPiece(pieceType, pieceColor, toSq)
	case move.Attack: pos.doAttack(pieceType, pieceColor, toSq)
	case move.WhiteAttackEP: pos.doAttack(pieceType, White, toSq-8)
	case move.BlackAttackEP: pos.doAttack(pieceType, Black, toSq+8)
	case move.PromoQ: pos.putPiece(Queen, pieceColor, toSq)
	case move.PromoR: pos.putPiece(Rook, pieceColor, toSq)
	case move.PromoB: pos.putPiece(Bishop, pieceColor, toSq)
	case move.PromoN: pos.putPiece(Knight, pieceColor, toSq)
	case move.PromoAttkQ: pos.doPromoAttack(Queen, pieceColor, toSq)
	case move.PromoAttkR: pos.doPromoAttack(Rook, pieceColor, toSq)
	case move.PromoAttkB: pos.doPromoAttack(Bishop, pieceColor, toSq)
	case move.PromoAttkN: pos.doPromoAttack(Knight, pieceColor, toSq)
	case move.WhiteCastleK: pos.doCastle(G1, H1, F1, White)
	case move.WhiteCastleQ: pos.doCastle(C1, A1, D1, White)
	case move.BlackCastleK: pos.doCastle(G8, H8, F8, Black)
	case move.BlackCastleQ: pos.doCastle(C8, A8, D8, Black)
	}

	if pieceType == Pawn {
		pos.HalfMove = 0
		if pieceColor == White && toSq-fromSq == 16 {
			pos.EPSq = toSq-8
		}
		if pieceColor == Black && fromSq-toSq == 16{
			pos.EPSq = toSq+8
		}
	}
	
	pos.Castling &= Spoilers[fromSq] & Spoilers[toSq]
	pos.Side ^= 1
	
	return pos
}

func (pos *Position) doPromoAttack(promoType, promoColor, toSq uint8) {
	attackedType, attackedColor := pos.getPieceOnSq(toSq)
	pos.removePiece(attackedType, attackedColor, toSq)
	pos.putPiece(promoType, promoColor, toSq)
	pos.HalfMove = 0
}

func (pos *Position) doAttack(attackerType, attackerColor, toSq uint8) {
	attackedType, attackedColor := pos.getPieceOnSq(toSq)
	pos.removePiece(attackedType, attackedColor, toSq)
	pos.putPiece(attackerType, attackerColor, toSq)
	pos.HalfMove = 0
}

func (pos *Position) doCastle(kingToSq, rookFromSq, rookToSq, color uint8) {
	pos.removePiece(Rook, color, rookFromSq)
	pos.putPiece(King, color, kingToSq)
	pos.putPiece(Rook, color, rookToSq)
}


func (pos *Position) putPiece(pieceType, pieceColor, sq uint8) {
	pos.Pieces[pieceType] = SetBit(pos.Pieces[pieceType], sq)
	pos.Colors[pieceColor] = SetBit(pos.Colors[pieceColor], sq)
}

func (pos *Position) removePiece(pieceType, pieceColor, sq uint8) {
	pos.Pieces[pieceType] = UnsetBit(pos.Pieces[pieceType], sq)
	pos.Colors[pieceColor] = UnsetBit(pos.Colors[pieceColor], sq)
}

func (pos *Position) getPieceOnSq(sq uint8) (pieceType, pieceColor uint8) {
	empty := ^(pos.Colors[White] | pos.Colors[Black])

	pawnType := ((pos.Pieces[Pawn] & SquareBB[sq]) >> sq) * Pawn
	knightType := ((pos.Pieces[Knight] & SquareBB[sq]) >> sq) * Knight
	bishopType := ((pos.Pieces[Bishop] & SquareBB[sq]) >> sq) * Bishop
	rookType := ((pos.Pieces[Rook] & SquareBB[sq]) >> sq) * Rook
	queenType := ((pos.Pieces[Queen] & SquareBB[sq]) >> sq) * Queen
	kingType := ((pos.Pieces[King] & SquareBB[sq]) >> sq) * King
	noType := ((empty & SquareBB[sq]) >> sq) * NoType

	whiteColor := ((pos.Colors[White] & SquareBB[sq]) >> sq) * White
	blackColor := ((pos.Colors[Black] & SquareBB[sq]) >> sq) * Black
	noColor := ((empty & SquareBB[sq]) >> sq) * NoColor

	return uint8(pawnType + knightType + bishopType + rookType + queenType + kingType + noType),
		uint8(whiteColor + blackColor + noColor)
}
