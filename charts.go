package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"maps-data-converter/model"
	"os"
	"regexp"
	"strings"
)

var reCharts = regexp.MustCompile(`^./maps/([^/]+)/[^/]+/([^/]+)/(([^/]+)\s+\(([A-Z]{4})\)|([^/]+))/([^/\n]+\..+)$`)

var remove = []string{
	"(", ")", "_", "-", "'", "INT", "Int", "Intl", "MCAS", "AB", "AAF",
	"Airstrips", "Airstrip", "Airport", "Range", "Highwaystrip", "Highway", "Strip", "Elite"}

func ParseCharts() error {
	file, err := os.Open("data/charts.txt")
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var charts []model.Chart
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := reCharts.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		station := strings.TrimSpace(matches[4])
		icao := matches[5]
		if icao == "" {
			station = matches[6]
		}
		key := buildKeyName(station)
		if key == "" {
			fmt.Printf("invalid chart: %s\n", line)
		}

		sanitizedCountry, code := sanitizeCountryName(matches[2])

		charts = append(charts, model.Chart{
			Key:     key,
			Theater: matches[1],
			Code:    code,
			Country: sanitizedCountry,
			Station: station,
			Icao:    icao,
			Path:    matches[2] + "/" + matches[3],
			File:    matches[7],
			Name:    sanitizeChartName(matches[3], matches[7]),
		})
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	output, err := os.Create("data/charts.json")
	if err != nil {
		return err
	}
	defer func(output *os.File) {
		_ = output.Close()
	}(output)

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(charts)
}

func buildKeyName(station string) string {
	for _, r := range remove {
		station = strings.ReplaceAll(station, r, "")
	}
	station = strings.ReplaceAll(station, " ", "")
	return strings.ToUpper(station)
}

var reSanitize = regexp.MustCompile(`^(?:\d{2}\s+)?((.+) \(([A-Za-z]+)\)|(.+))$`)

func sanitizeCountryName(country string) (string, string) {
	matches := reSanitize.FindStringSubmatch(country)
	if matches == nil {
		return country, ""
	}
	if matches[4] != "" {
		return matches[4], ""
	}
	return matches[2], matches[3]
}

func sanitizeChartName(path, filename string) string {
	name := strings.ReplaceAll(filename, path, "")
	name = TrimFromFirstDot(name)
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, " - ", " ")
	name = strings.TrimSpace(name)
	if name == "" {
		name = "noname"
		fmt.Printf("invalid chart: %s\n", filename)
	}
	return name
}

func TrimFromFirstDot(name string) string {
	if i := strings.IndexByte(name, '.'); i >= 0 {
		return name[:i]
	}
	return name
}
