package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"server-tic-tac/game"
	"server-tic-tac/player"
	"server-tic-tac/room"
	"sync"
)

var waitingRoom []*player.Player
var mu sync.RWMutex

func WSHandler(ws *websocket.Conn) {
	ws.MaxPayloadBytes = 1024
	//defer ws.Close()

	//Store connected user to ActiveClients
	p := new(player.Player)
	p.Conn = ws
	p.Vals = []int{}
	mu.Lock()
	waitingRoom = append(waitingRoom, p)
	mu.Unlock()
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

		//watch if connection is ALIVE
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
		go room.StartMatch(p, p1)
	}
}

func main() {
	port := "9876"

	http.HandleFunc("/", func(writer http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(writer, "<p>This is a socket game server. Dial ws://%s:9876/ws </p>", r.URL.Host)
	})
	http.Handle("/ws", websocket.Handler(WSHandler))

	log.Println("Server listening at http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
