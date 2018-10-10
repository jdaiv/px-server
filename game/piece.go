package game

const (
	INVALID = iota
	KING
	QUEEN
	BISHOP
	ROOK
	KNIGHT
	PAWN
)

type Piece struct {
	Board *Board
	Owner *Player
	Type  int
}

func CreatePiece(owner *Player, pieceType int) Piece {
	return Piece{
		Owner: owner,
		Type:  pieceType,
	}
}

func (p *Piece) ValidMove(board Board, fromX, fromY, toX, toY int) bool {
	return true
	/*
		if !board.CheckBounds(toX, toY) {
			return false
		} else if p.Type == KING {
			return false
		} else if p.Type == QUEEN {
			return false
		} else if p.Type == BISHOP {
			return false
		} else if p.Type == ROOK {
			return false
		} else if p.Type == KNIGHT {
			return false
		} else if p.Type == PAWN {
			return false
		} else {
			return false
		}
	*/
}
