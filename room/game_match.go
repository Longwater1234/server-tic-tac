package room

import (
	"golang.org/x/net/websocket"
	"log"
	"server-tic-tac/game"
	"server-tic-tac/player"
	"strconv"
)

// StartMatch and update results until either disconnects or gameOver
func StartMatch(p2 *player.Player, p1 *player.Player) {
	// default starts with X (player 1)
	var isPlayerXTurn = true
	//keeps record of the player (playerName -> []indexes)
	var playerState = map[string]*player.Player{
		p1.Name: p1,
		p2.Name: p2,
	}

	for {
		var payload game.Payload
		if isPlayerXTurn {
			isPlayerXTurn = true
			if err := websocket.JSON.Receive(p1.Conn, &payload); err != nil {
				log.Printf("%s disconnected", p1.Name)
				p1.Conn.Close()
				e := p2.SendMessage(&game.Payload{
					MessageType: game.EXIT,
					Content:     "OPPONENT LEFT GAME",
					FromUser:    p1.Name,
				})
				handleError(p2, e)
				//p2.Conn.Close()
				return
			}
		} else {
			isPlayerXTurn = false
			if err := websocket.JSON.Receive(p2.Conn, &payload); err != nil {
				log.Printf("%s disconnected", p2.Name)
				p2.Conn.Close()
				e := p1.SendMessage(&game.Payload{
					MessageType: game.EXIT,
					Content:     "OPPONENT LEFT GAME",
					FromUser:    p2.Name,
				})
				handleError(p1, e)
				//p1.Conn.Close()
				return
			}
		}

		switch payload.MessageType {
		case game.MOVE:
			/* Record move FROM player [X or 0];
			* Example payload >> p1{MOVE, "9", "X"} or p2{MOVE, "4", "O"}
			* Check winner
			* Notify both players. Repeat until Winner or draw
			 */
			var p = playerState[payload.FromUser]
			log.Printf("Player %v moved to index %v", p.Name, payload.Content)
			gridIndex, _ := strconv.Atoi(payload.Content)
			p.Vals = append(p.Vals, gridIndex)
			if p.HasWon() {
				e := p.SendMessage(&game.Payload{
					MessageType: game.WIN,
					Content:     "Congrats! You won! GAME OVER",
				})
				handleError(p, e)
				var loser = p2
				if !isPlayerXTurn {
					loser = p1
				}
				e = loser.SendMessage(&game.Payload{
					MessageType: game.LOSE,
					Content:     "Sorry! You lost! GAME OVER",
				})
				handleError(p, e)
				return
			}
		default:
			log.Println("Unknown command sent. Closing connections")
			p1.Conn.Close()
			p2.Conn.Close()
			return
		}
	}
}

func handleError(p *player.Player, err error) {
	log.Printf("Player %v left. Error: %s", p.Name, err.Error())
	if err != nil {
		p.Conn.Close()
	}
}
