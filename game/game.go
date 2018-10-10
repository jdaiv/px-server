package game

const (
	NORTH = iota
	EAST
	SOUTH
	WEST
)

type Game struct {
	Board     *Board
	PlayerOne *Player
	PlayerTwo *Player
}

type Player struct {
	Id   string
	Side int
}

func CreateGame(playerOne, playerTwo string) Game {
	return Game{
		Board: CreateBoard(10, 10),
		PlayerOne: &Player{
			Id:   playerOne,
			Side: SOUTH,
		},
		PlayerTwo: &Player{
			Id:   playerTwo,
			Side: NORTH,
		},
	}
}
