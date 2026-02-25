package main

import (
	"encoding/json"
	"fmt"
	"maps-data-converter/model"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type NominatimSearchResult struct {
	DisplayName string `json:"display_name"`
}

// EnrichStationsWithCountry fetches the country name for a given list of stations
// based on the first part of the station's name (the city).
// It first tries to find a mapping in country_mapping.json.
// It returns a slice of strings containing the names of stations for which no country was found.
func EnrichStationsWithCountry(theater string, stations []model.Station) []string {
	var missing []string
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	mappingFile := fmt.Sprintf("data/%s/country_mapping.json", theater)
	mapping := loadCountryMapping(mappingFile)
	updated := false

	for i := range stations {
		station := &stations[i]

		// Try mapping first
		if country, ok := mapping[station.Name]; ok {
			station.Country = country
			continue
		}

		city := extractCity(station.Name)
		if city == "" {
			missing = append(missing, station.Name)
			continue
		}

		fmt.Printf("Fetching country for %s (%s)\n", station.Name, city)
		country, err := fetchCountryForCity(client, city)
		if err != nil {
			fmt.Printf("Error fetching country for %s (%s): %v\n", station.Name, city, err)
			missing = append(missing, station.Name)
			continue
		}

		if country == "" {
			missing = append(missing, station.Name)
		} else {
			station.Country = country
			mapping[station.Name] = country
			updated = true
		}

		// Nominatim Usage Policy: 1 request per second
		time.Sleep(1100 * time.Millisecond)
	}

	if updated {
		if err := saveCountryMapping(mappingFile, mapping); err != nil {
			fmt.Printf("Warning: could not save updated country mapping: %v\n", err)
		}
	}

	return missing
}

func loadCountryMapping(path string) map[string]string {
	mapping := make(map[string]string)
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Warning: could not load country mapping from %s: %v\n", path, err)
		return mapping
	}

	if err := json.Unmarshal(data, &mapping); err != nil {
		fmt.Printf("Warning: could not parse country mapping from %s: %v\n", path, err)
	}

	return mapping
}

func saveCountryMapping(path string, mapping map[string]string) error {
	data, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func extractCity(name string) string {
	// The first part of the station name is normally the city
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == ' ' || r == '/' || r == '-' || r == ','
	})
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func fetchCountryForCity(client *http.Client, city string) (string, error) {
	apiURL := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1&accept-language=en", url.QueryEscape(city))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	// Nominatim requires a User-Agent
	req.Header.Set("User-Agent", "Maps-Data-Converter/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var results []NominatimSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", nil
	}

	// The last part of display_name is the country
	parts := strings.Split(results[0].DisplayName, ",")
	if len(parts) == 0 {
		return "", nil
	}

	country := strings.TrimSpace(parts[len(parts)-1])
	return country, nil
}
