package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	table := tview.NewTable().
		SetBorders(true)
	
	table.SetCell(0, 0, tview.NewTableCell("Time")).
		SetCell(0, 1, tview.NewTableCell("Temp (°C)")).
		SetCell(0, 2, tview.NewTableCell("Wind"))
	
	city := flag.String("city", "", "Name of the city")
	lat := flag.Float64("lat", 0.0, "Latitude of the location")
	lon := flag.Float64("lon", 0.0, "Longitude of the location")
	hours := flag.Int("hours", 6, "Number of hours for the forecast")
	
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

	forecast, err := fetchForecast(latitude, longitude, *hours)
	if err != nil {
		fmt.Println("Error fetching forecast:", err)
		return
	}
	fmt.Printf("Current temperature: %.1f°C\n", forecast.CurrentWeather.Temperature)
	cardinal := degreesToCardinal9(forecast.CurrentWeather.WindDirection)
	fmt.Printf("Wind: %.1f km/h, direction: %s\n", forecast.CurrentWeather.WindSpeed, cardinal)
	fmt.Printf("Time: %s\n", forecast.CurrentWeather.Time)
	hoursToShow := *hours // from your --hours flag

	count := 0 // how many hours printed
	for i := 0; i < len(forecast.Hourly.Time) && count < hoursToShow; i++ {
    t, err := time.Parse("2006-01-02T15:04", forecast.Hourly.Time[i])
    if err != nil {
        fmt.Println("Error parsing time:", err)
        continue
    }

    localTime := t.Local()
    if localTime.Before(now) {
        continue
    }

    timeStr := localTime.Format("15:04")
    tempStr := fmt.Sprintf("%.1f", forecast.Hourly.Temperature2m[i])
    
    // Wind if you have wind arrays
	// If wind arrays exist, print wind, else skip
	windStr := "N/A"
	if forecast.Hourly.WindDir10m != nil && forecast.Hourly.Windspeed10m != nil && len(forecast.Hourly.WindDir10m) > i && len(forecast.Hourly.Windspeed10m) > i {
		windDir := degreesToCardinal9(forecast.Hourly.WindDir10m[i])
		windSpeed := forecast.Hourly.Windspeed10m[i]
		windStr = fmt.Sprintf("%s %.1f km/h", windDir, windSpeed)
	}
	table.SetCell(count+1, 0, tview.NewTableCell(timeStr))
	table.SetCell(count+1, 1, tview.NewTableCell(tempStr))
	table.SetCell(count+1, 2, tview.NewTableCell(windStr))
	count++
}
// Run the TUI once after filling the table
	if err := app.SetRoot(table, true).Run(); err != nil {
		panic(err)
	}
}

func degreesToCardinal9(degrees float64) string {
	directions := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW", "N"}
	index := int((degrees + 22.5) / 45.0) % 8
	return directions[index]
}
type GeoResponse struct {
	Results []struct {
		Name 	 string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Country   string  `json:"country"`
	} `json:"results"`
}

type ForecastResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		WindSpeed   float64 `json:"windspeed"`
		WindDirection float64 `json:"winddirection"`
		WeatherCode  int     `json:"weathercode"`
		Time         string  `json:"time"`
	} `json:"current_weather"`
	Hourly struct {
		Time            []string  `json:"time"`
		Temperature2m   []float64 `json:"temperature_2m"`
		Precipitation   []float64 `json:"precipitation"`
		WindDir10m      []float64 `json:"winddirection_10m"`
		Windspeed10m    []float64 `json:"windspeed_10m"`
	} `json:"hourly"`
}

func fetchForecast(lat, lon float64, hours int) (*ForecastResponse, error) {
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&hourly=temperature_2m,precipitation&current_weather=true&timezone=auto", lat, lon)
	resp , err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body , err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var forecast ForecastResponse
	err  = json.Unmarshal(body, &forecast)
	if err != nil {
		return nil, err
	}
	return &forecast, nil
}

func geocodeCity(city string) (float64, float64, error) {
	url := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1", city)
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var geo GeoResponse
	err = json.Unmarshal(body, &geo)
	if err != nil {
		return 0, 0, err
	}

	if len(geo.Results) == 0 {
		return 0, 0, fmt.Errorf("no results found for city: %s", city)
	}

	return geo.Results[0].Latitude, geo.Results[0].Longitude, nil
}
