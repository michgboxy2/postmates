package ws

import (
	"fmt"
	"sync"
)

type Hub struct {
	clients         map[string]*WSConnection
	broadcast       chan []byte
	Register        chan *WSConnection
	unregister      chan *WSConnection
	rooms           map[string]*Room
	createRoomQueue chan RoomRequest
	joinRoomQueue   chan RoomRequest
	leaveRoomQueue  chan RoomRequest
	gil             sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		Register:   make(chan *WSConnection),
		unregister: make(chan *WSConnection),

		clients: make(map[string]*WSConnection),

		rooms:           make(map[string]*Room),
		createRoomQueue: make(chan RoomRequest),
		joinRoomQueue:   make(chan RoomRequest),
		leaveRoomQueue:  make(chan RoomRequest),

		gil: sync.Mutex{},
	}
}

func (h *Hub) GetSize() int {
	h.gil.Lock()
	defer h.gil.Unlock()
	return len(h.clients)
}

func (h *Hub) GetCourier(id string) *WSConnection {
	return h.GetClient(fmt.Sprintf("courier_%s", id))
}

func (h *Hub) GetCustomer(id uint) *WSConnection {
	return h.GetClient(fmt.Sprintf("customer_%d", id))
}

func (h *Hub) GetClient(id string) *WSConnection {
	h.gil.Lock()
	defer h.gil.Unlock()
	return h.clients[id]
}

func (h *Hub) createRoom(name string) {
	if _, roomExists := h.rooms[name]; roomExists {
		return
	}

	h.rooms[name] = NewRoom(name)
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.Register:
			h.gil.Lock()
			h.clients[conn.GetIdBasedOnType()] = conn

			h.gil.Unlock()
		case conn := <-h.unregister:
			h.gil.Lock()
			if _, ok := h.clients[conn.GetIdBasedOnType()]; ok {
				delete(h.clients, conn.GetIdBasedOnType())
				conn.Deactivate()
			}
			h.gil.Unlock()

		case request := <-h.createRoomQueue:
			h.createRoom(request.name)

		case request := <-h.joinRoomQueue:
			room := h.rooms[request.name]
			room.joinQueue <- request

		case request := <-h.leaveRoomQueue:
			room := h.rooms[request.name]
			room.leaveQueue <- request
		}
	}
}
