package osrm

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

type Profile string

const (
	Driving Profile = "driving"
)

type Route struct {
	Code   string `json:"code"`
	Routes []struct {
		Duration float64 `json:"duration"`
	} `json:"routes"`
}

type StringLonLat string

type GetTravelTimeInput struct {
	StartLonLat StringLonLat `json:"start_lon_lat"`
	EndLonLat   StringLonLat `json:"end_lon_lat"`
	Profile     Profile      `json:"profile"`
}

func (s StringLonLat) Location() (Location, error) {
	temp := strings.Split(string(s), ",")
	if len(temp) != 2 {
		return Location{}, fmt.Errorf("expected lon,lat but instead received %v", s)

	}
	lat, err := strconv.ParseFloat(temp[1], 64)
	if err != nil {
		return Location{}, fmt.Errorf("unable to parse latitude from %v", temp[1])
	}
	long, err := strconv.ParseFloat(temp[0], 64)
	if err != nil {
		return Location{}, fmt.Errorf("unable to parse longitude from %v", temp[0])
	}
	return Location{
		Latitude:  lat,
		Longitude: long,
	}, nil
}

func (i GetTravelTimeInput) Execute() (float64, error) {
	start, err := i.StartLonLat.Location()
	if err != nil {
		return 0, err
	}
	end, err := i.EndLonLat.Location()
	if err != nil {
		return 0, err
	}
	profile := i.Profile
	if profile == "" {
		profile = "driving"
	}
	return GetTravelTime(start, end, i.Profile)
}

func GetTravelTime(start, end Location, profile Profile) (float64, error) {
	if profile == "" {
		profile = "driving"
	}
	url := fmt.Sprintf(
		`https://router.project-osrm.org/route/v1/%v/%v,%v;%v,%v`,
		profile,
		start.Longitude,
		start.Latitude,
		end.Longitude,
		end.Latitude,
	)
	log.Println(url)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return 0, fmt.Errorf("unexpected status from OSRM: %v", response.Status)
	}

	var route Route
	err = json.NewDecoder(response.Body).Decode(&route)
	if err != nil {
		return 0, err
	}
	if route.Code != "Ok" {
		return 0, fmt.Errorf("unexpected code in JSON from OSRM: %v", route.Code)
	}
	if len(route.Routes) != 1 {
		return 0, fmt.Errorf("expected to receive a single route from OSRM but instead received: %v", route.Routes)
	}
	return route.Routes[0].Duration, nil
}
