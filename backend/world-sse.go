package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const patience time.Duration = time.Millisecond * 20

// Broker keeps list of open clients and brodcast events.
// Broker holds an instance of a World.
//
type Broker struct {
	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte
	// New client connections
	newClients chan chan []byte
	// Closed client connections
	closingClients chan chan []byte
	// Client connections registry
	clients map[chan []byte]bool
	// World
	world *World
}

// NewServer creates a broker instance and starts a new
// go routine to listen all client actions.
//
func NewServer() (broker *Broker) {
	broker = &Broker{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
		world:          NewWorld(),
	}
	go broker.listen()
	return
}

// sendConnectionID sends a newConnection event to current connection.
//
func (broker *Broker) sendConnectionID(w http.ResponseWriter) {

	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	newConnection := struct {
		ID int `json:"id"`
	}{
		len(broker.clients),
	}

	if b, err := json.Marshal(newConnection); err == nil {
		fmt.Fprintf(w, "event:newConnection\ndata:%s\n\n", b)
	}

	flusher.Flush()
}

func (broker *Broker) sendWorldUpdate(w http.ResponseWriter, time []byte) {

	// Make sure that the writer supports flushing.
	//
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	n := len(broker.clients)
	randID := 1
	if n != 0 {
		randID = rand.Intn(n)
	}

	var entity Entity
	entity = broker.world.MoveEntity(randID)

	data := struct {
		ID   int    `json:"id"`
		Type string `json:"type"`
		Time string `json:"time"`
		X    int    `json:"x"`
		Y    int    `json:"y"`
	}{
		randID,
		"position",
		fmt.Sprintf("%s", time),
		entity.x,
		entity.y,
	}

	if b, err := json.Marshal(data); err == nil {
		fmt.Fprintf(w, "data:%s\n\n", b)
	}
	flusher.Flush()
}

// ServeHTTP handles an HTTP request for broker server send requests.
//
func (broker *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	messageChan := make(chan []byte)
	broker.newClients <- messageChan

	broker.sendConnectionID(w)

	// Remove this client from the map of connected clients when this handler exits.
	defer func() {
		broker.closingClients <- messageChan
	}()

	// Listen to connection close and un-register messageChan
	notify := w.(http.CloseNotifier).CloseNotify()

	for {
		select {
		case <-notify:
			return
		default:
			broker.sendWorldUpdate(w, <-messageChan)
		}
	}
}

// listen listens all client actions in broker.
//
func (broker *Broker) listen() {
	for {
		select {
		case s := <-broker.newClients:

			// A new client has connected.
			// Register their message channel
			broker.clients[s] = true
			log.Printf("Client added. %d registered clients", len(broker.clients))
		case s := <-broker.closingClients:

			// A client has dettached and we want to
			// stop sending them messages.
			delete(broker.clients, s)
			log.Printf("Removed client. %d registered clients", len(broker.clients))
		case event := <-broker.Notifier:

			// We got a new event from the outside!
			// Send event to all connected clients
			for clientMessageChan := range broker.clients {
				select {
				case clientMessageChan <- event:
				case <-time.After(patience):
					log.Print("Skipping client.")
				}
			}
		}
	}
}
