package main

type Hub struct {
	clients  map[string]*Client
	rooms    map[string]*Room
	register chan struct {
		roomID string
		client *Client
	}
	unregister chan string
}

func newHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
		rooms:   make(map[string]*Room),
		register: make(chan struct {
			roomID string
			client *Client
		}),
		unregister: make(chan string),
	}
}

// run: run async
func (hub *Hub) run() {
	for {
		select {
		case reg := <-hub.register:
			room, ok := hub.rooms[reg.roomID]
			if !ok {
				room = newRoom(reg.roomID)
				room.unregisterToHubCh = hub.unregister
				go room.run()
				hub.rooms[reg.roomID] = room
			}

			// if more than two clients connect, close them all
			if len(room.clients) == 2 {
				reg.client.close()
				room.unregisterCh <- struct{}{}
				continue
			}

			room.registerCh <- reg.client

		case id := <-hub.unregister:
			delete(hub.rooms, id)
			// fmt.Printf("Room %s removed\n", id)
		}
	}
}
