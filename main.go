package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {

	city := flag.String("city", "", "Name of the city")
	lat := flag.Float64("lat", 0.0, "Latitude of the location")
	lon := flag.Float64("lon", 0.0, "Longitude of the location")
	hours := flag.Int("hours", 24, "Number of hours for the forecast")
	
	flag.Parse()

	if *city == "" && (*lat == 0.0 || *lon == 0.0) {
		fmt.Println("Error: You must provide either --city or both --lat and --lon")
		flag.Usage()
		return
	}
	var latitude, longitude float64

	if *city != "" {
		var err error
		latitude, longitude, err = geocodeCity(*city)
		if err != nil {
			fmt.Println("Error geocoding city:", err)
			return
		}
	}else {
		latitude = *lat
		longitude = *lon
	}
	now := time.Now()
	fmt.Println("City:", *city)					
	fmt.Println("Latitude:", latitude)
	fmt.Println("Longitude:", longitude)
	fmt.Println("Hours:", *hours)
	fmt.Println("Current Time:", now.Format("2006-01-02 15:04:05"))
}

type GeoResponse struct {
	Results []struct {
		Name 	 string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Country   string  `json:"country"`
	} `json:"results"`
}

type forecastResponse struct {
	latitude  float64 `json:"latitude"`
	longitude float64 `json:"longitude"`
	
func geocodeCity(city string) (float64, float64, error) {
	apiURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1", city)
	resp, err := http.Get(apiURL)
	if err != nil {
		return 0.0, 0.0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0.0, 0.0, err
	}
	var geo GeoResponse
	if err := json.Unmarshal(body, &geo); err != nil {
		return 0.0, 0.0, err
	}
	if len(geo.Results) == 0 {
		return 0.0, 0.0, fmt.Errorf("no results found for city: %s", city)
	}
	return geo.Results[0].Latitude, geo.Results[0].Longitude, nil
}