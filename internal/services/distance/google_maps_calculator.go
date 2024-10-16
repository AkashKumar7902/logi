package distance

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "logi/internal/models"
    "net/http"
    "net/url"
    "time"
)

// GoogleMapsCalculator implements the DistanceCalculator interface using Google Maps API.
type GoogleMapsCalculator struct {
    APIKey string
    Client *http.Client
}

// NewGoogleMapsCalculator returns a new instance of GoogleMapsCalculator.
func NewGoogleMapsCalculator(apiKey string) *GoogleMapsCalculator {
    return &GoogleMapsCalculator{
        APIKey: apiKey,
        Client: &http.Client{Timeout: 10 * time.Second},
    }
}

// Calculate computes the distance and duration using Google Maps Distance Matrix API.
func (g *GoogleMapsCalculator) Calculate(pickup, dropoff models.Location) (*DistanceResult, error) {
    origin := fmt.Sprintf("%f,%f", pickup.Coordinates[1], pickup.Coordinates[0])
    destination := fmt.Sprintf("%f,%f", dropoff.Coordinates[1], dropoff.Coordinates[0])

    endpoint := "https://maps.googleapis.com/maps/api/distancematrix/json"
    params := url.Values{}
    params.Add("origins", origin)
    params.Add("destinations", destination)
    params.Add("key", g.APIKey)
    params.Add("units", "metric")

    reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

    resp, err := g.Client.Get(reqURL)
    if err != nil {
        log.Printf("Google Maps API request failed: %v", err)
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Printf("Google Maps API returned status: %s", resp.Status)
        return nil, errors.New("failed to fetch distance from Google Maps API")
    }

    var result GoogleMapsResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        log.Printf("Failed to decode Google Maps API response: %v", err)
        return nil, err
    }

    if result.Status != "OK" {
        log.Printf("Google Maps API error: %s", result.Status)
        return nil, errors.New("error from Google Maps API")
    }

    if len(result.Rows) == 0 || len(result.Rows[0].Elements) == 0 {
        return nil, errors.New("no results from Google Maps API")
    }

    element := result.Rows[0].Elements[0]
    if element.Status != "OK" {
        log.Printf("Google Maps API element error: %s", element.Status)
        return nil, errors.New("error in distance element from Google Maps API")
    }

    distanceKm := float64(element.Distance.Value) / 1000.0 // meters to kilometers
    durationMin := float64(element.Duration.Value) / 60.0    // seconds to minutes

    return &DistanceResult{
        Distance: distanceKm,
        Duration: durationMin,
    }, nil
}

// GoogleMapsResponse represents the JSON response from Google Maps Distance Matrix API.
type GoogleMapsResponse struct {
    Status string `json:"status"`
    Rows   []struct {
        Elements []struct {
            Status   string `json:"status"`
            Distance struct {
                Text  string `json:"text"`
                Value int    `json:"value"` // in meters
            } `json:"distance"`
            Duration struct {
                Text  string `json:"text"`
                Value int    `json:"value"` // in seconds
            } `json:"duration"`
        } `json:"elements"`
    } `json:"rows"`
}
