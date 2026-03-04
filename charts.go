package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"maps-data-converter/model"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var reCharts = regexp.MustCompile(`^./maps/([^/]+)/[^/]+/([^/]+)/(([^/]+)\s+\(([A-Z]{4})\)|([^/]+))/([^/\n]+\..+)$`)

var remove = []string{
	"(", ")", "_", "-", "'", "Intl", "INT", "Int", "MCAS", "AB", "AAF",
	"Airstrips", "Airstrip", "Airport", "Airbase", "Range", "Highwaystrip", "Highway", "Strip", "Elite"}

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
	name = trimExt(name)
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, " - ", " ")
	name = strings.TrimSpace(name)
	if name == "" {
		name = "noname"
		fmt.Printf("invalid chart: %s\n", filename)
	}
	return name
}

func trimExt(name string) string {
	if i := strings.IndexByte(name, '.'); i >= 0 {
		return name[:i]
	}
	return name
}

func GenerateChartFiles(chartsJSONPath, outDir string) error {
	f, err := os.Open(chartsJSONPath)
	if err != nil {
		return fmt.Errorf("error opening charts.json: %w", err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	var rows []model.Chart
	dec := json.NewDecoder(f)
	if err := dec.Decode(&rows); err != nil {
		return fmt.Errorf("error while parsing charts.json: %w", err)
	}

	grouped := make(map[string]map[string][]model.ChartRef)

	for _, r := range rows {
		if r.Key == "" {
			fmt.Printf("ignoring chart without key: %s\n", r.Name)
			continue
		}

		name := r.Name
		url := r.Path + "/" + r.File

		if grouped[r.Theater] == nil {
			grouped[r.Theater] = make(map[string][]model.ChartRef)
		}
		grouped[r.Theater][r.Key] = append(grouped[r.Theater][r.Key], model.ChartRef{
			Name: name,
			URL:  url,
		})
	}

	theaters := make([]string, 0, len(grouped))
	for t := range grouped {
		theaters = append(theaters, t)
	}
	sort.Strings(theaters)

	for _, theater := range theaters {
		theaterDir := filepath.Join(outDir, theater)
		if err := os.MkdirAll(theaterDir, 0o755); err != nil {
			return fmt.Errorf("theater-Verzeichnis anlegen %q: %w", theaterDir, err)
		}

		keys := make([]string, 0, len(grouped[theater]))
		for k := range grouped[theater] {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			refs := grouped[theater][key]

			// Optional: deterministisch nach Name+URL sortieren
			sort.SliceStable(refs, func(i, j int) bool {
				if refs[i].Name == refs[j].Name {
					return refs[i].URL < refs[j].URL
				}
				return refs[i].Name < refs[j].Name
			})

			outFile := filepath.Join(theaterDir, key+".json")
			b, err := json.MarshalIndent(refs, "", "  ")
			if err != nil {
				return fmt.Errorf("error serializing json for %s/%s: %w", theater, key, err)
			}
			b = append(b, '\n')

			if err := os.WriteFile(outFile, b, 0o644); err != nil {
				return fmt.Errorf("error writing file %q: %w", outFile, err)
			}
		}
	}

	return nil
}
