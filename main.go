package main

import (
	"fmt"
	"maps-data-converter/model"
	"math"
)

const MapFeet = 3359580.0 // 3358699.5 // Falcon const
const BmsMapScale = 1 / MapFeet * 1024000

const MapPixels = 4096

func main() {

	basePath := "c:\\apps\\Falcon BMS 4.38"

	ctFile := basePath + "\\Data\\TerrData\\Objects\\Falcon4_CT.xml"
	campFile := basePath + "\\Data\\Campaign\\CampObjData.XML"
	stationFile := basePath + "\\Data\\Campaign\\Stations+Ils.dat"

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

	stations, err := ParseStationFreqFile(stationFile)
	stationMap := make(map[string]*model.StationFreq, len(stations))
	for _, st := range stations {
		if st == nil {
			continue
		}
		stationMap[st.Key] = st
	}

	if err != nil {
		panic(err)
	}
	fmt.Printf("Station ILS records count:  %d\n", len(stations))

	for _, station := range stations {
		fmt.Printf("Station: %s, Key: %s\n", station.Name, station.Key)
	}

	var records []model.Station

	for _, abObj := range *airbaseObjects {
		abRecord := CreateAirbaseRecord(&abObj)

		abKey := getKey(abRecord.Name)

		phdPath, pdxPath := CreateDataPaths(abRecord.OcdIdx)

		phd, err := loader.LoadPHD(basePath + phdPath)
		if err != nil {
			panic(err)
		}

		pdx, err := loader.LoadPDX(basePath + pdxPath)
		if err != nil {
			panic(err)
		}

		ExtractRunwayData(phd.PHDs, pdx.PDs, abRecord)

		fmt.Printf("%s: %s, Pos x:%.5f y:%.5f, OcdIdx:%d\n", abKey, abRecord.Name, abRecord.Pos.X, abRecord.Pos.Y, abRecord.OcdIdx)

		fac := MapFeet / MapPixels

		px := int(math.Round(abRecord.Pos.X / fac))
		py := int(math.Round(MapPixels - abRecord.Pos.Y/fac))
		fmt.Printf("Map-Pos x:%d y:%d\n", px, py)

		for idx, rw := range abRecord.Runways {
			fmt.Printf("Runway %d, Length %f, Width %f, Heading %f\n", idx, rw.Length, rw.Width, rw.Heading)
		}

		detail := model.Details{
			Name: abRecord.Name,
			Lat:  "",
			Long: "",
			Elev: "",
			Rwy:  BuildRunwayInfo(abRecord.Runways),
		}

		freq := stationMap[abKey]
		if freq == nil {
			fmt.Printf("No station data for %s\n", abKey)
		} else {
			fmt.Printf("Station name %s, %s\n", freq.Name, freq.Atis)
			detail.Twr = freq.TowerUhf
			detail.Atis = freq.Atis
			detail.Gnd = freq.Ground
			detail.Ops = freq.Ops
			detail.Tcn = freq.Tacan
			detail.AppDep = freq.Approach
		}

		sta := model.Station{
			Name:    abKey,
			Country: "KTO",
			Type:    "Airbase",
			PosX:    px,
			PosY:    py,
			Details: &detail,
		}

		records = append(records, sta)
	}

	err = writeRecordsJSON("stations.json", records)
	if err != nil {
		panic(err)
	}

}
