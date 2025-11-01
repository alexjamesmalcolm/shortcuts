package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"shortcuts/job"
	"shortcuts/osrm"
	"shortcuts/route"
	"strconv"
)

func main() {
	job.StartTaskMaster()
	job.DefineJob[osrm.GetTravelTimeInput]("/run/travel-time/")
	job.DefineJob[route.OptimalRouteInput]("/run/optimal-route/")

	portString := os.Getenv("PORT")
	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		log.Fatalf(`Tried to parse uint from environment variable PORT but was unable to parse "%v"`, portString)
	}
	log.Printf("Listening on port %v\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}
