package room

import (
	"log"
	"server-tic-tac/game"
	"server-tic-tac/player"
	"strconv"

	"golang.org/x/net/websocket"
)

// StartMatch and keep updating results until either disconnects or gameOver
func StartMatch(p1 *player.Player, p2 *player.Player, gameOver chan bool) {
	log.Println("🟢 Match has begun!")

	p1.SendMessage(&game.Payload{
		MessageType: game.START,
		Content:     "Make your first move",
		FromUser:    player.X.String(),
	})

	p2.SendMessage(&game.Payload{
		MessageType: game.START,
		Content:     "Match has started. PlayerX's turn",
		FromUser:    player.O.String(),
	})

	// default starts with X (player 1)
	var isPlayerXTurn = true

	for {
		if isPlayerXTurn {
			//IT'S PLAYER 1'S TURN
			var payload game.Payload
			if err := websocket.JSON.Receive(p1.Conn, &payload); err != nil {
				log.Println(p1.Name, "disconnected. Cause:", err.Error())
				p2.SendMessage(&game.Payload{
					MessageType: game.EXIT,
					Content:     "OPPONENT LEFT GAME",
					FromUser:    p1.Name,
				})
				gameOver <- true
				return
			}
			log.Println("Payload:", payload)

			//FORWARD THE "MOVE" PAYLOAD TO PLAYER 2, FOR UI UPDATE
			if err := websocket.JSON.Send(p2.Conn, &payload); err != nil {
				log.Println(p2.Name, "disconnected. Cause %+v", err.Error())
				p1.SendMessage(&game.Payload{
					MessageType: game.EXIT,
					Content:     "OPPONENT LEFT GAME",
					FromUser:    p2.Name,
				})
				gameOver <- true
				return
			}

			//RECORD THE MOVE, CHECK WINNER or DRAW
			if checkEndGame(p1, p2, &payload) {
				gameOver <- true
				return
			}
			isPlayerXTurn = false
		} else {
			//IT'S PLAYER 2'S TURN
			var payload game.Payload
			if err := websocket.JSON.Receive(p2.Conn, &payload); err != nil {
				log.Println(p2.Name, "disconnected. Cause:", err.Error())
				p1.SendMessage(&game.Payload{
					MessageType: game.EXIT,
					Content:     "OPPONENT LEFT GAME",
					FromUser:    p2.Name,
				})
				gameOver <- true
				return
			}
			log.Println("Payload:", payload)

			//FORWARD the "MOVE" payload TO PLAYER 1, FOR UI UPDATE
			if err := websocket.JSON.Send(p1.Conn, &payload); err != nil {
				log.Println(p1.Name, "disconnected. Cause:", err.Error())
				p2.SendMessage(&game.Payload{
					MessageType: game.EXIT,
					Content:     "OPPONENT LEFT GAME",
					FromUser:    p1.Name,
				})
				gameOver <- true
				return
			}

			//RECORD THE MOVE, CHECK WINNER or DRAW
			if checkEndGame(p2, p1, &payload) {
				gameOver <- true
				return
			}
			isPlayerXTurn = true
		}
	}
}

// Records the move by Player `p` against `opponent`, then returns match status, whether it's Over
func checkEndGame(p, opponent *player.Player, payload *game.Payload) bool {
	gridIndex, _ := strconv.Atoi(payload.Content)
	p.Cells = append(p.Cells, gridIndex)
	if checkWinner(p, opponent) {
		return true
	}
	if checkDraw(p, opponent) {
		return true
	}
	return false
}

// checks if `p` has won against `opponent`. if TRUE, notify both.
func checkWinner(p, opponent *player.Player) bool {
	if ok, winningCells := p.HasWon(); ok {
		p.SendMessage(&game.Payload{
			MessageType:  game.WIN,
			Content:      "Congrats! You won! GAME OVER",
			WinningCells: winningCells,
		})
		opponent.SendMessage(&game.Payload{
			MessageType:  game.LOSE,
			Content:      "Sorry! You lost! GAME OVER",
			WinningCells: winningCells,
		})
		log.Println("🏆 We got a winner!", p.Name, "has won the match")
		return true
	}
	return false
}

// checks if the match is in a Draw. if TRUE, notify both.
func checkDraw(p, opponent *player.Player) bool {
	totalUsed := len(p.Cells) + len(opponent.Cells)
	if totalUsed == 9 {
		//It's a draw!
		p.SendMessage(&game.Payload{
			MessageType: game.DRAW,
			Content:     "It's a draw! GAME OVER",
		})
		opponent.SendMessage(&game.Payload{
			MessageType: game.DRAW,
			Content:     "It's a draw! GAME OVER",
		})
		log.Println("😑 It's a Draw!")
		return true
	}
	return false
}
