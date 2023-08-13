package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"server-tic-tac/game"
	"server-tic-tac/player"
	"server-tic-tac/room"
	"sync/atomic"

	"golang.org/x/net/websocket"
)

var numPlayers atomic.Uint32             // total number of LIVE players
var lobby = make(chan *player.Player, 2) // waiting room for players

// wsHandler handles 1 client per connection
func wsHandler(ws *websocket.Conn) {
	ws.MaxPayloadBytes = 1024
	var clientIp = ws.Request().RemoteAddr
	defer ws.Close()

	p := &player.Player{
		Conn:  ws,
		Cells: []int{},
		Dead:  make(chan bool, 1),
	}
	defer close(p.Dead)

	//for each pair joining, the 1st will be player X
	if numPlayers.Load()%2 == 0 {
		p.Name = player.X.String()
	} else {
		p.Name = player.O.String()
	}
	numPlayers.Add(1)
	lobby <- p

	log.Println("Someone connected", clientIp, "Total players:", numPlayers.Load())
	<-p.Dead
	numPlayers.Add(^uint32(0)) // minus 1
	log.Println(clientIp, p.Name, "just left the game. Total players:", numPlayers.Load())
}

func main() {
	port := "9876"
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)

	http.HandleFunc("/", func(writer http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(writer, `<p>This is a socket game server. Dial ws://%s:%s/ws </p>`, r.URL.Host, port)
	})
	http.Handle("/ws", websocket.Handler(wsHandler))

	go listenForJoins()
	log.Println("Server listening at http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Listen for new players joining lobby
func listenForJoins() {
	for {
		log.Println("LOBBY:", "cap", cap(lobby), "len", len(lobby))
		p1 := <-lobby
		p1.SendMessage(&game.Payload{
			MessageType: game.WELCOME,
			Content:     "Connected. Waiting for opponent...",
			FromUser:    player.X.String(),
		})
		p2 := <-lobby
		p2.SendMessage(&game.Payload{
			MessageType: game.WELCOME,
			Content:     "Connected. Game is started!",
			FromUser:    player.O.String(),
		})

		//start the match in new goroutine
		go func() {
			gameOver := make(chan bool, 1)
			room.StartMatch(p1, p2, gameOver)
			//block until match ends
			<-gameOver
			log.Println("🔴 GAME OVER!")
			p1.Dead <- true
			p2.Dead <- true
		}()
	}
}
