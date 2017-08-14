// This package implements an HTTP server providing a REST API for ....
//
//
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Println("starting server at 8081..")

	broker := NewServer()

	go func() {
		for {
			time.Sleep(time.Millisecond * 20)
			eventString := fmt.Sprintf("%v", time.Now())
			// log.Println("Receiving event")
			broker.Notifier <- []byte(eventString)
		}
	}()

	http.HandleFunc("/api/sse", broker.ServeHTTP)
	http.ListenAndServe(":8081", nil)
}
