package chattingroom

import (
	"log"
	"net/http"
)

var (
	globalHub Hub
	inited    = false
)

func init() {
	if !inited {
		globalHub = Hub{
			broadcast:    make(chan Message),
			register:     make(chan *Connection),
			unregister:   make(chan *Connection),
			rooms:        make(map[string]map[*Connection]bool),
			connReadonly: true,
		}

		inited = true
	}
}

// Hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	// Registered connections.
	rooms map[string]map[*Connection]bool

	// Inbound messages from the connections.
	broadcast chan Message

	// Register requests from the connections.
	register chan *Connection

	// Unregister requests from connections.
	unregister chan *Connection

	connReadonly bool
}

func GetHub() *Hub {
	return &globalHub
}

func (h *Hub) Rooms() map[string]map[*Connection]bool {
	return h.rooms
}

func (h *Hub) SetConnReadonly(b bool) {
	h.connReadonly = b
}

func (h *Hub) Broadcast(msg Message) {
	h.broadcast <- msg
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			connections := h.rooms[c.room]
			if connections == nil {
				connections = make(map[*Connection]bool)
				h.rooms[c.room] = connections
			}
			h.rooms[c.room][c] = true
		case c := <-h.unregister:
			connections := h.rooms[c.room]
			if connections != nil {
				if _, ok := connections[c]; ok {
					delete(connections, c)
					close(c.send)
					if len(connections) == 0 {
						delete(h.rooms, c.room)
					}
				}
			}
		case m := <-h.broadcast:
			connections := h.rooms[m.room]
			for c := range connections {
				select {
				case c.send <- m.data:
				default:
					close(c.send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(h.rooms, m.room)
					}
				}
			}
		}
	}
}

func (h *Hub) Handle(room string, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	c := &Connection{
		send:     make(chan []byte, 256),
		ws:       ws,
		room:     room,
		readonly: h.connReadonly,
	}
	h.register <- c
	go c.writePump()
	c.readPump()
}
