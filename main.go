package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var hub *Hub

func init() {
	hub = newHub()
	go hub.run()
}

var port = flag.String("port", ":3000", "")

func main() {
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		vals := r.URL.Query()
		temp, ok := vals["room"]
		if !ok {
			fmt.Fprintf(w, "You need to supply the room")
			return
		}
		roomID := temp[0]

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Err: ", err)
			return
		}

		client := newClient(conn)
		client.run()

		hub.register <- struct {
			roomID string
			client *Client
		}{roomID, client}

	})

	fmt.Println("Running on port:", (*port)[1:])
	http.ListenAndServe(*port, nil)
}
