package room

import (
	"golang.org/x/net/websocket"
	"log"
	"server-tic-tac/game"
	"server-tic-tac/player"
	"strconv"
)

// StartMatch and update results until either disconnects or gameOver
func StartMatch(p1 *player.Player, p2 *player.Player, done chan bool) {
	p1.Name = "X"
	p2.Name = "O"
	eee := p1.SendMessage(&game.Payload{
		MessageType: game.START,
		Content:     "Make your move",
		FromUser:    player.X.String(),
	})
	if eee != nil {
		log.Println("first", eee)
		done <- true
	}

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
				log.Printf("%s disconnected. Error :%v", p1.Name, err.Error())
				p1.Conn.Close()
				p2.SendMessage(&game.Payload{
					MessageType: game.EXIT,
					Content:     "OPPONENT LEFT GAME",
					FromUser:    p1.Name,
				})
				done <- true
				return
			}
		} else {
			isPlayerXTurn = false
			if err := websocket.JSON.Receive(p2.Conn, &payload); err != nil {
				log.Printf("%s disconnected", p2.Name)
				p2.Conn.Close()
				p1.SendMessage(&game.Payload{
					MessageType: game.EXIT,
					Content:     "OPPONENT LEFT GAME",
					FromUser:    p2.Name,
				})
				done <- true
				return
			}
		}

		switch payload.MessageType {
		case game.MOVE:
			/* Record move FROM player [X or 0];
			* Example payload >> p1{MOVE, "9", "X"} or p2{MOVE, "4", "O"},
			* AS {messageType, Content, FromPlayer}
			 */
			var p = playerState[payload.FromUser]
			log.Printf("Player %v moved to index %v", p.Name, payload.Content)
			gridIndex, _ := strconv.Atoi(payload.Content)
			p.Vals = append(p.Vals, gridIndex)

			//forward to opponent
			var opponent = p2
			if !isPlayerXTurn {
				opponent = p1
			}
			opponent.SendMessage(&payload)

			//check winner
			if p.HasWon() {
				e := p.SendMessage(&game.Payload{
					MessageType: game.WIN,
					Content:     "Congrats! You won! GAME OVER",
				})
				handleError(p, e, done)
				e = opponent.SendMessage(&game.Payload{
					MessageType: game.LOSE,
					Content:     "Sorry! You lost! GAME OVER",
				})
				handleError(p, e, done)
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

func handleError(p *player.Player, err error, done chan bool) {
	log.Printf("Player %v left. Error: %s", p.Name, err.Error())
	if err != nil {
		done <- true
		p.Conn.Close()
	}
}
