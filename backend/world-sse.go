package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const patience time.Duration = time.Millisecond * 20

// Broker keeps list of open clients and brodcast events.
// Broker holds an instance of a World.
//
type Broker struct {
	Notifier       chan []byte
	newClients     chan chan []byte
	closingClients chan chan []byte
	clients        map[chan []byte]bool

	newClientStats     chan chan int
	closingClientStats chan chan int
	clientStats        map[chan int]bool

	NotifierRemove        chan []byte
	newClientRemoves      chan chan []byte
	closingClientsRemoves chan chan []byte
	clientRemoves         map[chan []byte]bool

	world         *World
	connectionIDs map[string]bool
	mutex         sync.Mutex
}

// NewServer creates a broker instance and starts a new
// go routine to listen all client actions.
//
func NewServer() (broker *Broker) {
	broker = &Broker{
		Notifier:              make(chan []byte, 1),
		newClients:            make(chan chan []byte),
		closingClients:        make(chan chan []byte),
		clients:               make(map[chan []byte]bool),
		newClientStats:        make(chan chan int),
		closingClientStats:    make(chan chan int),
		clientStats:           make(map[chan int]bool),
		NotifierRemove:        make(chan []byte, 1),
		newClientRemoves:      make(chan chan []byte),
		closingClientsRemoves: make(chan chan []byte),
		clientRemoves:         make(map[chan []byte]bool),
		world:                 NewWorld(),
		connectionIDs:         make(map[string]bool),
	}
	go broker.listen()
	go broker.listenStats()
	go broker.listenRemoves()
	return
}

func (broker *Broker) storeConnectionID(connectionID string) {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()

	if _, ok := broker.connectionIDs[connectionID]; !ok {
		broker.connectionIDs[connectionID] = true
	}
}

func (broker *Broker) removeConnectionID(connectionID string) {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()
	delete(broker.connectionIDs, connectionID)
}

func (broker *Broker) randConnectionID() string {
	broker.mutex.Lock()
	defer broker.mutex.Unlock()
	n := rand.Intn(len(broker.connectionIDs))
	for id := range broker.connectionIDs {
		if n == 0 {
			return id
		}
		n--
	}
	return ""
}

// sendConnectionID sends a newConnection event to current connection.
//
func (broker *Broker) sendConnectionID(w http.ResponseWriter, connectionID string) {

	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	newConnection := struct {
		ID string `json:"id"`
	}{
		connectionID,
	}

	if b, err := json.Marshal(newConnection); err == nil {
		fmt.Fprintf(w, "event:newConnection\ndata:%s\n\n", b)
	}

	flusher.Flush()
}

// sendConnectionID sends a newConnection event to current connection.
//
func (broker *Broker) sendRemoveConnectionID(w http.ResponseWriter, connectionID string) {

	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	connection := struct {
		ID string `json:"id"`
	}{
		connectionID,
	}

	if b, err := json.Marshal(connection); err == nil {
		fmt.Fprintf(w, "event:removeConnection\ndata:%s\n\n", b)
	}

	flusher.Flush()
}

func (broker *Broker) sendNumberOfClientsUpdate(w http.ResponseWriter, numClients int) {
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	data := struct {
		NumClients int `json:"numClients"`
	}{
		numClients,
	}

	if b, err := json.Marshal(data); err == nil {
		fmt.Fprintf(w, "event:numClients\ndata:%s\n\n", b)
	}
	flusher.Flush()
}

func (broker *Broker) sendWorldUpdate(w http.ResponseWriter, time []byte) {

	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	randConnectionID := broker.randConnectionID()

	var entity Entity
	entity = broker.world.MoveEntity(randConnectionID)

	data := struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Time string `json:"time"`
		X    int    `json:"x"`
		Y    int    `json:"y"`
	}{
		randConnectionID,
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

	statsChan := make(chan int)
	broker.newClientStats <- statsChan

	removeChan := make(chan []byte)
	broker.newClientRemoves <- removeChan

	// Each connection registers its own message channel with the Broker's connections registry
	messageChan := make(chan []byte)
	broker.newClients <- messageChan

	connectionID := fmt.Sprintf("%s", time.Now())

	broker.storeConnectionID(connectionID)
	broker.sendConnectionID(w, connectionID)

	// Remove this client from the map of connected clients when this handler exits.
	defer func() {
		log.Printf("closing chans")
		broker.closingClients <- messageChan
		broker.closingClientStats <- statsChan
		broker.closingClientsRemoves <- removeChan
		broker.removeConnectionID(connectionID)
		broker.NotifierRemove <- []byte(connectionID)

	}()

	// Listen to connection close and un-register messageChan
	notify := w.(http.CloseNotifier).CloseNotify()

	for {
		select {
		case <-notify:
			return
		case stats := <-statsChan:
			log.Printf("sendNumberOfClientsUpdate")
			broker.sendNumberOfClientsUpdate(w, stats)
		case id := <-removeChan:
			broker.sendRemoveConnectionID(w, string(id[:]))
		default:
			// log.Printf("sendWorldUpdate")
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

func (broker *Broker) listenRemoves() {
	for {
		select {
		case s := <-broker.newClientRemoves:
			broker.clientRemoves[s] = true
		case event := <-broker.NotifierRemove:
			for clientMessageChan := range broker.clientRemoves {
				select {
				case clientMessageChan <- event:
				case <-time.After(patience):
					log.Print("Skipping client.")
				}
			}
		case s := <-broker.closingClientsRemoves:
			delete(broker.clientRemoves, s)
			log.Printf("Removed client. %d registered removes clients", len(broker.clientRemoves))
		}
	}
}

// listenStats listens all client stat actions in broker.
//
func (broker *Broker) listenStats() {
	for {
		select {
		case s := <-broker.newClientStats:
			broker.clientStats[s] = true
			log.Printf("NumberOfClients client added. %d registered clients", len(broker.clientStats))
			for msgChan := range broker.clientStats {
				log.Printf("New client stats >>>")
				msgChan <- len(broker.clients)
			}
		case s := <-broker.closingClientStats:
			delete(broker.clientStats, s)
			log.Printf("Removed number of clients. %d registered clients", len(broker.clientStats))
		}
	}
}
