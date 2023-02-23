package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"server-tic-tac/game"
	"server-tic-tac/player"
	"strconv"
	"sync"
)

var waitingRoom []*player.Player
var mu sync.RWMutex

func WSHandler(ws *websocket.Conn) {
	ws.MaxPayloadBytes = 1024
	defer ws.Close()

	//Store connected user to ActiveClients
	p := new(player.Player)
	p.Conn = ws
	p.Vals = []int{}
	waitingRoom = append(waitingRoom, p)
	log.Printf("Someone connected")

	if len(waitingRoom)%2 != 0 {
		//currentPlayer := "X"
		p.Name = player.X.String()
		p.SendMessage(&game.Payload{
			MessageType: game.WELCOME,
			Content:     "Welcome. You are player X. Waiting for opponent",
			FromUser:    player.X.String(),
		})
		log.Printf("Player X ready. Waiting for opponent")
	} else {
		p.Name = player.O.String()
		p.SendMessage(&game.Payload{
			MessageType: game.WELCOME,
			Content:     "Welcome. You are player O. Found opponent. Starting game",
			FromUser:    player.O.String(),
		})
		log.Printf("Player O now ready. Starting game")
		mu.Lock()
		p1 := waitingRoom[len(waitingRoom)-2]
		waitingRoom = waitingRoom[:len(waitingRoom)-2]
		mu.Unlock()
		go startMatch(p, p1)
	}
}

func startMatch(p2 *player.Player, p1 *player.Player) {
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
				p2.SendMessage(&game.Payload{
					MessageType: game.EXIT,
					Content:     "OPPONENT LEFT GAME",
					FromUser:    p1.Name,
				})
				//p2.Conn.Close()
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
				p.SendMessage(&game.Payload{
					MessageType: game.WIN,
					Content:     "Congrats! You won! GAME OVER",
				})
				var playerLoser = p2
				if !isPlayerXTurn {
					playerLoser = p1
				}
				playerLoser.SendMessage(&game.Payload{
					MessageType: game.LOSE,
					Content:     "Sorry! You lost! GAME OVER",
				})
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

func main() {
	port := "9876"

	http.HandleFunc("/", func(writer http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(writer, "<p>This is a socket game server. Dial http://%s:9876/ws </p>", r.URL.Host)
	})
	http.Handle("/ws", websocket.Handler(WSHandler))

	log.Println("Server listening at http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
