package main

import (
	"fmt"
	"maps-data-converter/model"
	"math"
	"sort"
	"strings"
)

func BuildRunwayInfo(runways []*model.Runway) string {
	if len(runways) == 0 {
		return ""
	}

	type item struct {
		idx int
		num int // 1..36
	}

	items := make([]item, 0, len(runways))
	for i, r := range runways {
		items = append(items, item{idx: i, num: headingToRunwayNumber(r.Heading)})
	}

	// Gruppen nach Runway-Nummer (für parallele Bahnen)
	groups := map[int][]item{}
	for _, it := range items {
		groups[it.num] = append(groups[it.num], it)
	}

	// Innerhalb der Gruppe stabil nach Input-Reihenfolge (optional, aber deterministisch)
	for k := range groups {
		sort.SliceStable(groups[k], func(i, j int) bool {
			return groups[k][i].idx < groups[k][j].idx
		})
	}

	// Gruppen sortieren: nach der kleineren der beiden Richtungsnummern
	keys := make([]int, 0, len(groups))
	for num := range groups {
		keys = append(keys, num)
	}
	sort.Slice(keys, func(i, j int) bool {
		ai := min(keys[i], reciprocalRunwayNumber(keys[i]))
		aj := min(keys[j], reciprocalRunwayNumber(keys[j]))
		if ai != aj {
			return ai < aj
		}
		return keys[i] < keys[j]
	})

	var parts []string
	for _, num := range keys {
		g := groups[num]
		suffixes := suffixesForCount(len(g))

		for i := range g {
			suf := suffixes[i]
			parts = append(parts, formatPairSmallFirst(num, suf))
		}
	}

	return strings.Join(parts, " - ")
}

func formatPairSmallFirst(num int, suf string) string {
	recip := reciprocalRunwayNumber(num)
	recipSuf := reciprocalSuffix(suf)

	// Wenn reciprocal kleiner ist, drehen wir die Darstellung um
	if recip < num {
		return fmt.Sprintf("%02d%s/%02d%s", recip, recipSuf, num, suf)
	}
	return fmt.Sprintf("%02d%s/%02d%s", num, suf, recip, recipSuf)
}

func headingToRunwayNumber(headingDeg float64) int {
	h := math.Mod(headingDeg, 360.0)
	if h < 0 {
		h += 360.0
	}
	n := int(math.Round(h / 10.0))
	if n == 0 {
		n = 36
	}
	if n > 36 {
		n = n % 36
		if n == 0 {
			n = 36
		}
	}
	return n
}

func reciprocalRunwayNumber(n int) int {
	r := n + 18
	if r > 36 {
		r -= 36
	}
	return r
}

func suffixesForCount(count int) []string {
	switch count {
	case 1:
		return []string{""}
	case 2:
		// Ohne Geometrie/Position ist L/R nicht sicher; wir bleiben bei einer festen,
		// reproduzierbaren Zuordnung.
		return []string{"R", "L"}
	case 3:
		return []string{"L", "C", "R"}
	default:
		out := make([]string, count)
		return out
	}
}

func reciprocalSuffix(s string) string {
	switch s {
	case "L":
		return "R"
	case "R":
		return "L"
	case "C":
		return "C"
	default:
		return ""
	}
}
