package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"maps-data-converter/model"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var stationRE = regexp.MustCompile(`^(?P<camp_id>\d+)\s(?P<tacan>\d+)\s(?P<band>[XY])\s\d+\s(?P<range>\d+)\s\d\s(?P<twr_uhf>\d+)\s(?P<twr_vhf>\d+)\s(?P<rwy1_ils>\d+)\s(?P<rwy2_ils>\d+)\s(?P<rwy3_ils>\d+)\s(?P<rwy4_ils>\d+)\s(?P<ops>\d+)\s(?P<ground>\d+)\s(?P<approach>\d+)\s(?P<lso>\d+)\s(?P<atis>\d+)\s\d+\s\d+\s\d+\s-?\d+$`)

func ParseStationFreqBytes(data []byte) ([]*model.StationFreq, error) {
	return ParseStationFreqReader(bytes.NewReader(data))
}

func ParseStationFreqFile(filename string) ([]*model.StationFreq, error) {
	// the file to parse is Data/Campaign/Stations+Ils.dat
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	return ParseStationFreqReader(bufio.NewReader(f))
}

func ParseStationFreqReader(r io.Reader) ([]*model.StationFreq, error) {
	var (
		scanner = bufio.NewScanner(r)
		out     []*model.StationFreq
	)

	push := func(line string) {
		m := stationRE.FindStringSubmatch(line)
		if m == nil {
			return
		}

		idx := map[string]int{}
		for i, n := range stationRE.SubexpNames() {
			if n != "" {
				idx[n] = i
			}
		}

		get := func(key string) string {
			if i, ok := idx[key]; ok && i < len(m) {
				return m[i]
			}
			return ""
		}

		rwy := []string{}
		for _, k := range []string{"rwy1_ils", "rwy2_ils", "rwy3_ils", "rwy4_ils"} {
			v := get(k)
			if v != "" && v != "0" {
				rwy = append(rwy, formatFreq(v))
			}
		}

		tcn := get("tacan")
		band := get("band")
		if tcn == "001" {
			tcn = ""
			band = ""
		}

		campId, err := strconv.Atoi(get("camp_id"))
		if err != nil {
			fmt.Printf("failed to parse camp_id: %v", err)
		}

		rng, err := strconv.Atoi(get("range"))
		if err != nil {
			fmt.Printf("failed to parse range: %v", err)
		}

		out = append(out, &model.StationFreq{
			CampId:   campId,
			Range:    rng,
			Tacan:    tcn,
			Band:     band,
			TowerUhf: sanitizeFreq(get("twr_uhf")),
			TowerVhf: sanitizeFreq(get("twr_vhf")),
			Runways:  rwy,
			Ops:      sanitizeFreq(get("ops")),
			Ground:   sanitizeFreq(get("ground")),
			Approach: sanitizeFreq(get("approach")),
			Atis:     sanitizeFreq(get("atis")),
		})
	}

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			push(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
