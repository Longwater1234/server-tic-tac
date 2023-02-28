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

var GameMatch = make(chan *player.Player, 1)
var waitingRoom []*player.Player
var mu sync.RWMutex

func WSHandler(ws *websocket.Conn) {
	ws.MaxPayloadBytes = 1024
	defer ws.Close()
	done := make(chan bool)

	p := new(player.Player)
	p.Conn = ws
	p.Vals = []int{}
	//p.Name == len(gameMatch) % 2

	var pieceType string
	mu.RLock()
	if len(waitingRoom)%2 == 0 {
		pieceType = "X"
	} else {
		pieceType = "O"
	}
	mu.RUnlock()

	err1 := p.SendMessage(&game.Payload{
		MessageType: game.WELCOME,
		Content:     fmt.Sprintf("Welcome. You are player %s.", pieceType),
		FromUser:    pieceType,
	})
	if err1 != nil {
		log.Println("ERROR IN WELCOME MESSAGE", err1)
		ws.Close()
	}
	GameMatch <- p
	log.Printf("Someone connected. Waiting room count: %d, cap %d", len(GameMatch), cap(GameMatch))
	go func() {
		p := <-GameMatch
		mu.Lock()
		waitingRoom = append(waitingRoom, p)
		mu.Unlock()
		if len(waitingRoom) > 0 && len(waitingRoom)%2 == 0 {
			log.Println("two players joined. Starting game")
			log.Println("game starting")
			room.StartMatch(waitingRoom[0], waitingRoom[1], done)
		}
	}()
	<-done
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
