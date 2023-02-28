package room

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"server-tic-tac/game"
	"server-tic-tac/player"
	"strconv"
)

// StartMatch and update results until either disconnects or gameOver
func StartMatch(p1 *player.Player, p2 *player.Player, done chan bool) {
	var err error
	p1.Name = player.X.String()
	p2.Name = player.O.String()

	log.Printf("Address %+v", p1)

	err = p1.SendMessage(&game.Payload{
		MessageType: game.START,
		Content:     "Make your move",
		FromUser:    player.X.String(),
	})

	if err != nil {
		handleError(err, done)
		return
	}

	err = p2.SendMessage(&game.Payload{
		MessageType: game.START,
		Content:     "Match has started. PlayerX's turn",
		FromUser:    player.O.String(),
	})

	// default starts with X (player 1)
	var isPlayerXTurn = true

	//keeps record of the player (playerName -> []indexes)
	var playerState = map[string]*player.Player{
		p1.Name: p1,
		p2.Name: p2,
	}

	for {
	free:
		var payload game.Payload
		if isPlayerXTurn {
			if err := websocket.JSON.Receive(p1.Conn, &payload); err != nil {
				log.Printf("%s disconnected. Error :%v", p1.Name, err.Error())
				p1.Conn.Close()
				err = p2.SendMessage(&game.Payload{
					MessageType: game.EXIT,
					Content:     "OPPONENT LEFT GAME",
					FromUser:    p1.Name,
				})
				done <- true
				return
			}
			fmt.Printf("%+v\n", payload)
			if err := websocket.JSON.Send(p2.Conn, payload); err != nil {
				log.Printf("%s disconnected. Cause %+v", p2.Name, err.Error())
				p2.Conn.Close()
				p1.SendMessage(&game.Payload{
					MessageType: game.EXIT,
					Content:     "OPPONENT LEFT GAME",
					FromUser:    p2.Name,
				})
				done <- true
				return
			}
			isPlayerXTurn = false
		} else {
			if err := websocket.JSON.Receive(p2.Conn, &payload); err != nil {
				log.Printf("%s disconnected. Cause %+v", p2.Name, err.Error())
				p2.Conn.Close()
				err = p1.SendMessage(&game.Payload{
					MessageType: game.EXIT,
					Content:     "OPPONENT LEFT GAME",
					FromUser:    p2.Name,
				})
				done <- true
				return
			}
			fmt.Printf("%+v\n", payload)
			if err := websocket.JSON.Send(p1.Conn, payload); err != nil {
				log.Printf("%s disconnected. Cause %+v", p1.Name, err.Error())
				p1.Conn.Close()
				p2.SendMessage(&game.Payload{
					MessageType: game.EXIT,
					Content:     "OPPONENT LEFT GAME",
					FromUser:    p1.Name,
				})
				done <- true
				return
			}
			isPlayerXTurn = true
		}

		//RECORD THE MOVE
		log.Printf("I am free")
		var p = playerState[payload.FromUser]
		log.Printf("AddressTwo 2 %+v", p)
		gridIndex, _ := strconv.Atoi(payload.Content)
		p.Vals = append(p.Vals, gridIndex)
		if checkWinner(p2, p1, done) {
			done <- true
			return
		}
		goto free
	}
}

func handleError(err error, done chan bool) {
	if err != nil {
		log.Printf("%v", err)
		done <- true
	}
}

func checkWinner(p, opponent *player.Player, done chan bool) bool {
	if p.HasWon() {
		err := p.SendMessage(&game.Payload{
			MessageType: game.WIN,
			Content:     "Congrats! You won! GAME OVER",
		})
		handleError(err, done)
		err = opponent.SendMessage(&game.Payload{
			MessageType: game.LOSE,
			Content:     "Sorry! You lost! GAME OVER",
		})
		handleError(err, done)
		return true
	}
	return false

}
