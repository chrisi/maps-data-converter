package main

import (
	"fmt"
	"maps-data-converter/model"
	"math"
	"sort"
	"strings"
)

const MapFeet = 3359580.0 // 3358699.5 // Falcon const
const BmsMapScale = 1 / MapFeet * 1024000

const MapPixels = 4096

func main() {

	configs := []model.Config{
		{
			Theater:  "korea",
			BasePath: "c:\\apps\\Falcon BMS 4.38\\Data",
		},
		{
			Theater:  "balkans",
			BasePath: "c:\\apps\\Falcon BMS 4.38\\Data\\Add-On Balkans",
		},
		{
			Theater:  "israel",
			BasePath: "c:\\apps\\Falcon BMS 4.38\\Data\\Add-On Israel",
		},
		{
			Theater:  "hellas",
			BasePath: "c:\\apps\\Falcon BMS 4.38\\Data\\Add-On Hellas",
		},
	}

	configByTheater := make(map[string]model.Config, len(configs))
	for _, cfg := range configs {
		configByTheater[cfg.Theater] = cfg
	}

	cfg := configByTheater["israel"]

	ctFile := cfg.BasePath + "\\TerrData\\Objects\\Falcon4_CT.xml"
	campFile := cfg.BasePath + "\\Campaign\\CampObjData.XML"
	stationFile := cfg.BasePath + "\\Campaign\\Stations+Ils.dat"
	heighMapFile := cfg.BasePath + "\\Terrdata\\" + cfg.Theater + "\\NewTerrain\\Heightmaps\\Heightmap.raw"

	hmCfg := TheaterConfig{
		HeightmapBounds: Bounds{
			MinX: 0.0,
			MaxX: 3358699.5,
			MinY: 0.0,
			MaxY: 3358699.5,
		},
		HeightmapWidth:  32768,
		HeightmapHeight: 32768,
	}

	hmReader, err := NewHeightmapReader(heighMapFile, hmCfg)
	if err != nil {
		panic(err)
	}
	defer hmReader.Close()

	loader := DataLoader{}

	ctRecords, err := loader.LoadCTRecords(ctFile)
	if err != nil {
		panic(err)
	}
	fmt.Printf("All indices count:          %d\n", len(ctRecords.CTs))

	airbaseIndices := FilterCTsForAirbases(ctRecords.CTs)
	fmt.Printf("Airbase indices count:      %d\n", len(airbaseIndices))

	campObjData, err := loader.LoadCampObjData(campFile)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Campaign Objects count:     %d\n", len(campObjData.CampObjs))

	airbaseObjects := FilterCampObjsByAirbaseEntityIdx(campObjData, airbaseIndices)
	fmt.Printf("Airbase Objects count:      %d\n", len(*airbaseObjects))

	beaconObjects := FilterCampObjsByNavBeacon(campObjData)
	fmt.Printf("NavBeacon Objects count:    %d\n", len(*beaconObjects))

	stations, err := ParseStationFreqFile(stationFile)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Station ILS records count:  %d\n", len(stations))

	stationMap := make(map[int]*model.StationFreq, len(stations))
	for _, st := range stations {
		stationMap[st.CampId] = st
	}

	var records []model.Station

	for _, abObj := range *airbaseObjects {
		abRecord := CreateAirbaseRecord(&abObj)

		phdPath, pdxPath := CreateDataPaths(abRecord.OcdIdx)

		phd, err := loader.LoadPHD(cfg.BasePath + phdPath)
		if err != nil {
			panic(err)
		}

		pdx, err := loader.LoadPDX(cfg.BasePath + pdxPath)
		if err != nil {
			panic(err)
		}

		ExtractRunwayData(phd.PHDs, pdx.PDs, abRecord)
		fmt.Printf("%d: %s, Pos x:%.5f y:%.5f, OcdIdx:%d\n", abRecord.Id, abRecord.Name, abRecord.Pos.X, abRecord.Pos.Y, abRecord.OcdIdx)

		fac := MapFeet / MapPixels
		px := int(math.Round(abRecord.Pos.X / fac))
		py := int(math.Round(MapPixels - abRecord.Pos.Y/fac))
		fmt.Printf("Map-Pos x:%d y:%d\n", px, py)

		for idx, rw := range abRecord.Runways {
			fmt.Printf("Runway %d, Length %f, Width %f, Heading %f\n", idx, rw.Length, rw.Width, rw.Heading)
		}

		elevation, err := hmReader.GetElevation(abRecord.Pos.X, abRecord.Pos.Y)
		if err != nil {
			fmt.Printf("Error getting elevation for %s: %v\n", abRecord.Name, err)
		} else {
			fmt.Printf("Elevation: %.0f\n", elevation)
		}

		detail := model.Details{
			Name: abRecord.Name,
			Elev: fmt.Sprintf("%.0f ft", elevation),
			Rwy:  BuildRunwayInfo(abRecord.Runways),
		}

		freq := stationMap[abRecord.Id]
		if freq == nil {
			fmt.Printf("No station data for %d: %s\n", abRecord.Id, abRecord.Name)
		} else {
			fmt.Printf("Station name %s, %s\n", abRecord.Name, freq.Atis)
			if freq.Range > 0 {
				detail.Range = fmt.Sprintf("%d NM", freq.Range)
			}
			detail.Twr = freq.TowerUhf
			detail.Atis = freq.Atis
			detail.Gnd = freq.Ground
			detail.Ops = freq.Ops
			detail.Tcn = freq.Tacan + freq.Band
			detail.AppDep = freq.Approach
		}

		abKey := getKey(abRecord.Name)
		chartsDir := fmt.Sprintf("data/%s/charts/", cfg.Theater)
		charts, err := loader.LoadCharts(chartsDir, abKey)
		if err != nil {
			fmt.Printf("Error loading charts for %s: %v\n", abKey, err)
		} else if len(charts) > 0 {
			detail.Charts = charts
			fmt.Printf("Loaded %d charts for %s\n", len(charts), abKey)
		}

		tp := "Airbase"
		if strings.Contains(strings.ToLower(abRecord.Name), "strip") {
			tp = "Airstrip"
		}

		sta := model.Station{
			CampId:  abRecord.Id,
			OcdIdx:  abRecord.OcdIdx,
			Name:    abRecord.Name,
			Country: "n/a",
			Type:    tp,
			PosX:    px,
			PosY:    py,
			Details: &detail,
		}

		records = append(records, sta)
	}

	for _, nbObj := range *beaconObjects {
		nbRecord := CreateAirbaseRecord(&nbObj)

		fac := MapFeet / MapPixels

		px := int(math.Round(nbRecord.Pos.X / fac))
		py := int(math.Round(MapPixels - nbRecord.Pos.Y/fac))
		fmt.Printf("Map-Pos x:%d y:%d\n", px, py)

		detail := model.Details{
			Name: nbRecord.Name,
		}

		freq := stationMap[nbRecord.Id]
		if freq == nil {
			fmt.Printf("No station data for %d: %s\n", nbRecord.Id, nbRecord.Name)
		} else {
			fmt.Printf("Station name %s, %s\n", nbRecord.Name, freq.Atis)
			if freq.Range > 0 {
				detail.Range = fmt.Sprintf("%d NM", freq.Range)
			}
			detail.Tcn = freq.Tacan + freq.Band
		}

		tp := "n/a"

		if strings.Contains(nbRecord.Name, "VOR/DME") {
			tp = "VOR/DME"
		}
		if strings.Contains(nbRecord.Name, "VORTAC") {
			tp = "VORTAC"
		}

		sta := model.Station{
			CampId:  nbRecord.Id,
			OcdIdx:  nbRecord.OcdIdx,
			Name:    nbRecord.Name,
			Country: "n/a",
			Type:    tp,
			PosX:    px,
			PosY:    py,
			Details: &detail,
		}

		records = append(records, sta)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Name < records[j].Name
	})

	missing := EnrichStationsWithCountry(cfg.Theater, records)
	if len(missing) > 0 {
		fmt.Printf("Missing countries for stations: %v\n", missing)
	}

	err = writeRecordsJSON("stations_"+cfg.Theater+".json", records)
	if err != nil {
		panic(err)
	}

}
