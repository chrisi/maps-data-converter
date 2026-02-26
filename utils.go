package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func formatFreq(s string) string {
	if len(s) < 5 {
		return s
	}
	pf := ""
	if len(s) == 5 {
		pf = "0"
	}
	return s[:3] + "." + s[3:] + pf
}

func sanitizeFreq(s string) string {
	if s == "0" {
		return ""
	}
	return formatFreq(s)
}

var reName = regexp.MustCompile(`(?i)^(.*?)(?: Air.*| Highway.*| Strip.*| TAC.*| VOR.*| INT.*| RAF.*)?$`)

func getName(s string) string {
	m := reName.FindStringSubmatch(s)
	if m == nil {
		panic("cannot find airbase name")
	}
	return m[1]
}

func getKey(s string) string {
	n := strings.Replace(getName(s), " ", "-", -1)
	return strings.ToUpper(n)
}

func replaceABWithAirbase(s string) string {
	re := regexp.MustCompile(`(?i)\bAB\b`)
	return re.ReplaceAllString(s, "Airbase")
}

func writeRecordsJSON(path string, records any) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("datei erstellen: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	if err := enc.Encode(records); err != nil {
		return fmt.Errorf("json schreiben: %w", err)
	}
	return nil
}
