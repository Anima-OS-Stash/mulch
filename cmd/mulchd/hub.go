package main

import (
	"github.com/Xfennec/mulch"
)

// Hub structure allows multiple clients to receive messages
// from mulchd.
type Hub struct {
	clients    map[*HubClient]bool
	broadcast  chan *mulch.Message
	register   chan *HubClient
	unregister chan *HubClient
}

// HubClient describes a client of a Hub
type HubClient struct {
	Messages   chan *mulch.Message
	clientInfo string
	target     string
	trace      bool
	hub        *Hub
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*HubClient]bool),
		broadcast:  make(chan *mulch.Message),
		register:   make(chan *HubClient),
		unregister: make(chan *HubClient),
	}
}

// Run will start the Hub, allowing messages to be sent and received
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			// fmt.Printf("new client: %s\n", client.clientInfo)
		case client := <-h.unregister:
			// fmt.Printf("del client: %s\n", client.clientInfo)
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Messages)
			}
		case message := <-h.broadcast:
			// fmt.Printf("broadcasting\n")
			for client := range h.clients {
				if client.target != message.Target &&
					message.Target != mulch.MessageNoTarget &&
					client.target != mulch.MessageAllTargets {
					continue // not for this client
				}
				if message.Type == mulch.MessageTrace && client.trace == false {
					continue // this client don't want traces
				}

				// Here was a 'select' with 'default' case, where
				// the same things as 'h.unregister' were done.
				// It was source of some race conditions, and it seems
				// useless. Removed.
				client.Messages <- message
			}
		}
	}
}

// Broadcast send a message to all clients of the Hub
// (if the target matches)
func (h *Hub) Broadcast(message *mulch.Message) {
	h.broadcast <- message
}

// Register a new client of the Hub
// clientInfo is not currently used but is supposed to differentiate
// the client. Target may be mulch.MessageNoTarget.
func (h *Hub) Register(info string, target string, trace bool) *HubClient {
	client := &HubClient{
		Messages:   make(chan *mulch.Message),
		clientInfo: info,
		target:     target,
		trace:      trace,
		hub:        h,
	}
	h.register <- client
	return client
}

// Unregister the client from the Hub
func (hc *HubClient) Unregister() {
	hc.hub.unregister <- hc
}

// SetTarget allows the client to change (receiving) target
func (hc *HubClient) SetTarget(target string) {
	hc.target = target
}
