package main

type Room struct {
	id      string
	clients map[string]*Client

	broadcastCh chan struct {
		from    string
		message struct {
			messageType int
			payload     []byte
		}
	}
	registerCh   chan *Client
	unregisterCh chan struct{}

	// reference to hub's unregister
	unregisterToHubCh chan string
}

func newRoom(id string) *Room {
	return &Room{
		id:      id,
		clients: make(map[string]*Client),
		broadcastCh: make(chan struct {
			from    string
			message struct {
				messageType int
				payload     []byte
			}
		}),
		registerCh:   make(chan *Client),
		unregisterCh: make(chan struct{}),
	}
}

func (room *Room) run() {
	for {
		select {
		case client := <-room.registerCh:
			client.broadcastToRoomCh = room.broadcastCh
			client.unregisterToRoomCh = room.unregisterCh

			room.clients[client.id] = client

		case <-room.unregisterCh:
			// if one client unregisters, unregiser all clients
			for key, client := range room.clients {
				client.close()

				delete(room.clients, key)
			}

			room.unregisterToHubCh <- room.id
		case m := <-room.broadcastCh:
			for id, client := range room.clients {
				if id == m.from {
					continue
				}
				client.sendCh <- m.message
			}
		}
	}
}

// func (room *Room) addClient(client *Client) {
// 	client.broadcastToRoomCh = room.broadcastCh
// 	// room.clients = append(room.clients, client)
// 	room.clients[client.id] = client
// }
