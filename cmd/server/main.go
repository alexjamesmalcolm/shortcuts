package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Message struct {
	Message string `json:"message"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(Message{"Hello world!"})
		if err != nil {
			panic("IDK")
		}
		_, err = w.Write(data)
		if err != nil {
			panic("IDK")
		}
	})

	portString := os.Getenv("PORT")
	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		log.Fatalf(`Tried to parse uint from environment variable PORT but was unable to parse "%v"`, portString)
	}
	log.Printf("Listening on port %v\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
