package main

import (
	"maps-data-converter/model"
	"testing"
)

func TestExtractCity(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"Seoul", "Seoul"},
		{"Seoul Airbase", "Seoul"},
		{"Gimpo Int'l", "Gimpo"},
		{"Osan/K-55", "Osan"},
		{"Pohang-dong", "Pohang"},
		{"Kwangju (Gwangju)", "Kwangju"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := extractCity(tt.name)
			if actual != tt.expected {
				t.Errorf("extractCity(%q) = %q; want %q", tt.name, actual, tt.expected)
			}
		})
	}
}

func TestEnrichStationsWithCountry_Integration(t *testing.T) {
	// Only run if specifically requested to avoid hitting API during regular tests
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	stations := []model.Station{
		{Name: "Seoul"},
		{Name: "Kwaksan"},
		{Name: "NonExistentCity12345"},
	}

	missing := EnrichStationsWithCountry(stations)

	if stations[0].Country != "South Korea" {
		t.Errorf("Expected Seoul to be in South Korea, got %q", stations[0].Country)
	}
	if stations[1].Country != "North Korea" {
		t.Errorf("Expected Kwaksan to be in North Korea, got %q", stations[1].Country)
	}
	if len(missing) != 1 || missing[0] != "NonExistentCity12345" {
		t.Errorf("Expected 1 missing station (NonExistentCity12345), got %v", missing)
	}
}
