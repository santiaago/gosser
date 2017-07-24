// This package implements an HTTP server providing a REST API for ....
//
//
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

func (broker *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Make sure that the writer supports flushing.
	//
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	messageChan := make(chan []byte)

	// Signal the broker that we have a new connection
	id := len(broker.clients)
	log.Printf("ServerHTTP %d", id)
	broker.newClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits.
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
			n := len(broker.clients)
			randID := 1
			if n != 0 {
				randID = rand.Intn(n)
			}

			var dot dot
			dot = broker.world.MoveDot(randID)

			data := struct {
				ID   int    `json:"id"`
				Type string `json:"type"`
				Time string `json:"time"`
				X    int    `json:"x"`
				Y    int    `json:"y"`
			}{
				randID,
				"position",
				fmt.Sprintf("%s", <-messageChan),
				dot.x,
				dot.y,
			}

			if b, err := json.Marshal(data); err == nil {
				fmt.Fprintf(w, "data:%s\n\n", b)
			}

			// Flush the data immediatly instead of buffering it for later.
			flusher.Flush()
		}
	}
}

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

func main() {
	log.Println("starting server at 8081..")

	broker := NewServer()

	go func() {
		for {
			time.Sleep(time.Millisecond * 20)
			eventString := fmt.Sprintf("%v", time.Now())
			log.Println("Receiving event")
			broker.Notifier <- []byte(eventString)
		}
	}()

	http.HandleFunc("/api/sse", broker.ServeHTTP)
	http.ListenAndServe(":8081", nil)
}
