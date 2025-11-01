package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"shortcuts/job"
	"shortcuts/osrm"
	"strconv"
)

func main() {
	job.MakeJob[osrm.GetTravelTimeInput, float64]("/run/travel-time/")

	portString := os.Getenv("PORT")
	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		log.Fatalf(`Tried to parse uint from environment variable PORT but was unable to parse "%v"`, portString)
	}
	log.Printf("Listening on port %v\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
