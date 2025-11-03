package route

import (
	"log"
	"math"
	"shortcuts/osrm"
	"strings"
	"time"

	"github.com/ernestosuarez/itertools"
)

type Location struct {
	Address   string `json:"address"`
	Latitude  string `json:"lat"`
	Longitude string `json:"lon"`
}

func (l Location) Street() string {
	return strings.Split(l.Address, "\n")[0]
}

type OptimalRouteInput struct {
	Origin      Location     `json:"origin"`
	Destination Location     `json:"destination"`
	Stops       []Location   `json:"stops"`
	Profile     osrm.Profile `json:"profile"`
}

type Result struct {
	TravelTimes   map[string]map[string]float64
	BestRoute     []Location
	BestRouteTime float64
}

type travelTimeResult struct {
	start, end string
	duration   float64
	err        error
}

func getTravelTime(start, end Location, profile osrm.Profile) travelTimeResult {
	var travelTime float64
	var err error
	for range 3 {
		travelTime, err = osrm.GetTravelTime(osrm.Location{
			Latitude:  start.Latitude,
			Longitude: start.Longitude,
		}, osrm.Location{
			Latitude:  end.Latitude,
			Longitude: end.Longitude,
		}, profile)
		if err == nil {
			log.Printf("Travel time from %v to %v is %v", start.Street(), end.Street(), travelTime)
			break
		}
		log.Printf("retrying")
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		return travelTimeResult{err: err}
	}
	return travelTimeResult{start: start.Address, end: end.Address, duration: travelTime}
}

type TravelTimeMap map[string]map[string]float64

type OptimalRouteResult struct {
	TravelTimes   TravelTimeMap `json:"travel_times"`
	BestRoute     []Location    `json:"best_route"`
	BestRouteTime float64       `json:"best_route_time"`
}

func (i OptimalRouteInput) Execute() (OptimalRouteResult, error) {
	locationCount := len(i.Stops) + 2
	var addressToLocation = make(map[string]Location)
	var addresses = make([]string, 0, locationCount)
	addresses = append(addresses, i.Origin.Address)
	addressToLocation[i.Origin.Address] = i.Origin
	for _, l := range i.Stops {
		addresses = append(addresses, l.Address)
		addressToLocation[l.Address] = l
	}
	addresses = append(addresses, i.Destination.Address)
	addressToLocation[i.Destination.Address] = i.Destination

	var travelTimes = make(TravelTimeMap)
	for pair := range itertools.PermutationsStr(addresses, 2) {
		start := addressToLocation[pair[0]]
		end := addressToLocation[pair[1]]
		if start.Address == i.Destination.Address || end.Address == i.Origin.Address {
			continue
		}
		result := getTravelTime(start, end, i.Profile)

		if result.err != nil {
			return OptimalRouteResult{}, result.err
		}
		if travelTimes[result.start] == nil {
			travelTimes[result.start] = make(map[string]float64)
		}
		travelTimes[result.start][result.end] = result.duration
	}

	var bestTravelTime = math.MaxFloat64
	var bestRouteAddresses = make([]string, 0, locationCount)
	for route := range itertools.PermutationsStr(addresses, len(addresses)) {
		if route[0] != i.Origin.Address || route[len(addresses)-1] != i.Destination.Address {
			continue
		}
		var routeTravelTime float64
		for i := range len(route) - 1 {
			start := route[i]
			end := route[i+1]
			routeTravelTime += travelTimes[start][end]
		}
		if routeTravelTime < bestTravelTime {
			bestTravelTime = routeTravelTime
			bestRouteAddresses = route
		}
	}
	var bestRoute = make([]Location, len(bestRouteAddresses))
	for i, address := range bestRouteAddresses {
		bestRoute[i] = addressToLocation[address]
	}
	return OptimalRouteResult{
		TravelTimes:   travelTimes,
		BestRoute:     bestRoute,
		BestRouteTime: bestTravelTime,
	}, nil
}
