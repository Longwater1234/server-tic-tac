/*
 * Copyright (c) 2023, Davis Tibbz, MIT License.
 */

package player

import (
	"server-tic-tac/game"

	"golang.org/x/exp/slices"
	"golang.org/x/net/websocket"
)

var winningPatterns = [][]int{
	{0, 1, 2},
	{3, 4, 5},
	{6, 7, 8},
	{0, 3, 6},
	{1, 4, 7},
	{2, 5, 8},
	{0, 4, 8},
	{2, 4, 6},
}

// Player of the game, only 2 per game
type Player struct {
	Conn *websocket.Conn
	Name string // Name can only be X or O
	Vals []int  // cell indexes clicked by player
}

type SymbolGame int

const (
	O SymbolGame = iota
	X
)

func (s SymbolGame) String() string {
	switch s {
	case O:
		return "O"
	case X:
		return "X"
	}
	return "unknown"
}

// HasWon returns true if player has won.
func (p *Player) HasWon() bool {
	var markedCells = p.Vals
	if len(markedCells) < 3 {
		return false
	}

	for i := 0; i < len(winningPatterns); i++ {
		arr := winningPatterns[i]
		if slices.Contains(markedCells, arr[0]) && slices.Contains(markedCells, arr[1]) && slices.Contains(markedCells, arr[2]) {
			return true
		}
	}
	return false
}

// SendMessage to given player in JSON
func (p *Player) SendMessage(payload *game.Payload) error {
	err := websocket.JSON.Send(p.Conn, payload)
	return err
}
