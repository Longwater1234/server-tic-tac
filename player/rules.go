/*
 * Copyright (c) 2023, Davis Tibbz, MIT License.
 */

package player

import (
	"log"
	"server-tic-tac/game"

	"golang.org/x/exp/slices"
	"golang.org/x/net/websocket"
)

// all possible winning grid patterns
var winningPatterns = [][]int32{
	{0, 1, 2},
	{3, 4, 5},
	{6, 7, 8},
	{0, 3, 6},
	{1, 4, 7},
	{2, 5, 8},
	{0, 4, 8},
	{2, 4, 6},
}

// Player of the game, only 2 allowed per session
type Player struct {
	Conn  *websocket.Conn // client connection
	Name  string          // Name can only be X or O
	Cells []int32         // cell indexes used by this player
	Dead  chan bool       // whether player has disconnected
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

// HasWon returns true if Player has won. If YES, also return winning cells
func (p *Player) HasWon() (bool, []int32) {
	var markedCells = p.Cells
	if len(markedCells) < 3 {
		return false, []int32{}
	}

	for i := 0; i < len(winningPatterns); i++ {
		row := winningPatterns[i]
		if slices.Contains(markedCells, row[0]) && slices.Contains(markedCells, row[1]) && slices.Contains(markedCells, row[2]) {
			return true, row
		}
	}
	return false, []int32{}
}

// SendMessage in JSON to this player
func (p *Player) SendMessage(payload *game.Payload) {
	err := websocket.JSON.Send(p.Conn, payload)
	if err != nil {
		log.Printf("Failed to sendMessage to %s. Cause %+v", p.Name, err)
		p.Dead <- true
	}
}
