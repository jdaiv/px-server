package game

import (
	"errors"
)

var ErrorOutOfBounds = errors.New("out of bounds")
var ErrorTileNotFound = errors.New("tile not found")

type Board struct {
	Width  int
	Height int

	Tiles []*Tile
}

func CreateBoard(width, height int) *Board {
	b := Board{
		Width:  width,
		Height: height,
		Tiles:  make([]*Tile, width*height),
	}
	for i := 0; i < width*height; i++ {
		b.Tiles[i] = &Tile{
			Board: &b,
		}
	}
	return &b
}

func (b *Board) CheckBounds(x, y int) bool {
	return x >= 0 && x < b.Width && y >= 0 && y < b.Height
}

func (b *Board) GetTile(x, y int) (*Tile, error) {
	if !b.CheckBounds(x, y) {
		return nil, ErrorOutOfBounds
	}

	return b.Tiles[b.getTileIndex(x, y)], nil
}

func (b *Board) SetTile(tile *Tile, x, y int) error {
	if !b.CheckBounds(x, y) {
		return ErrorOutOfBounds
	}

	b.Tiles[b.getTileIndex(x, y)] = tile
	return nil
}

func (b *Board) getTileIndex(x, y int) int {
	return y*b.Width + x
}
