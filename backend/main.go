// This package implements an HTTP server providing a REST API for ....
//
//
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	log.Println("staring server at 8081..")

	http.HandleFunc("/api/hello", hello)
	http.ListenAndServe(":8081", nil)
}

func hello(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Message string `json:"message"`
	}{
		"hello",
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println(err)
		http.Error(w, "unable to encode data", http.StatusInternalServerError)
	}
}
