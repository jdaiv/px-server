package game

const (
	NORTH = iota
	EAST
	SOUTH
	WEST
)

type Game struct {
	Board  Board
	Pieces []Piece

	PlayerOne Player
	PlayerTwo Player
}

type Player struct {
	Id   string
	Name string
	Side int
}

func CreateGame() Game {
	return Game{
		Board:  CreateBoard(10, 10),
		Pieces: make([]Piece, 40),
		PlayerOne: Player{
			Id:   "One",
			Name: "One",
			Side: SOUTH,
		},
		PlayerTwo: Player{
			Id:   "Two",
			Name: "Two",
			Side: NORTH,
		},
	}
}
