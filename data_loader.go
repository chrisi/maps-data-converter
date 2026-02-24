package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"maps-data-converter/model"
	"math"
	"os"
	"strings"
)

type DataLoader struct {
}

func (dl DataLoader) LoadCTRecords(filePath string) (*model.CTRecords, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read xml file %q: %w", filePath, err)
	}

	var records model.CTRecords
	if err := xml.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("unmarshal ct records xml: %w", err)
	}

	return &records, nil
}

func (dl DataLoader) LoadCampObjData(filePath string) (*model.CampObjData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read xml file %q: %w", filePath, err)
	}

	var records model.CampObjData
	if err := xml.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("unmarshal campObj records xml: %w", err)
	}

	return &records, nil
}

func (dl DataLoader) LoadPHD(filePath string) (*model.PHDRecords, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read xml file %q: %w", filePath, err)
	}

	var records model.PHDRecords
	if err := xml.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("unmarshal PHD records xml: %w", err)
	}

	return &records, nil
}

func (dl DataLoader) LoadPDX(filePath string) (*model.PDRecords, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read xml file %q: %w", filePath, err)
	}

	var records model.PDRecords
	if err := xml.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("unmarshal PHX records xml: %w", err)
	}

	return &records, nil
}

func (dl DataLoader) LoadCharts(chartsDir, airbaseKey string) ([]*model.Chart, error) {
	filePath := chartsDir + "\\" + strings.ToUpper(airbaseKey) + ".json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read charts file %q: %w", filePath, err)
	}

	var charts []*model.Chart
	if err := json.Unmarshal(data, &charts); err != nil {
		return nil, fmt.Errorf("unmarshal charts json: %w", err)
	}

	return charts, nil
}

func FilterCTsForAirbases(in []model.CT) []model.CT {
	out := make([]model.CT, 0, len(in))
	for _, ct := range in {
		if ct.Domain != 3 {
			continue
		}
		if ct.Class != 4 {
			continue
		}
		if ct.Type != 1 && ct.Type != 2 {
			continue
		}
		if ct.Specific != 255 {
			continue
		}
		out = append(out, ct)
	}
	return out
}

func FilterCampObjsByAirbaseEntityIdx(campObjData *model.CampObjData, airbases []model.CT) *[]model.CampObj {
	airbaseIdx := make(map[int]struct{}, len(airbases))
	for _, ab := range airbases {
		airbaseIdx[ab.EntityIdx] = struct{}{}
	}

	out := make([]model.CampObj, 0)
	for _, obj := range campObjData.CampObjs {
		if _, ok := airbaseIdx[obj.OcdIndex]; ok {
			out = append(out, obj)
		}
	}
	return &out
}

func FilterCampObjsByNavBeacon(campObjData *model.CampObjData) *[]model.CampObj {
	out := make([]model.CampObj, 0)
	for _, obj := range campObjData.CampObjs {
		if obj.OcdIndex == 10 {
			out = append(out, obj)
		}
	}
	return &out
}

func CreateAirbaseRecord(ab *model.CampObj) *model.Airbase {
	return &model.Airbase{
		Id:   ab.CampId,
		Name: ab.CampName,
		Pos: model.Position{
			X: ab.PositionY,
			Y: ab.PositionX,
		},
		OcdIdx:  ab.OcdIndex,
		Runways: []*model.Runway{},
	}
}

func CreateDataPaths(ocdIdx int) (string, string) {
	basePath := "\\Data\\TerrData\\Objects\\ObjectiveRelatedData"
	return fmt.Sprintf("%s\\OCD_%05d\\PHD_%05d.xml", basePath, ocdIdx, ocdIdx),
		fmt.Sprintf("%s\\OCD_%05d\\PDX_%05d.xml", basePath, ocdIdx, ocdIdx)
}

const RUNWAY = 8

func ExtractRunwayData(phds []model.PHD, pdxs []model.PD, airbase *model.Airbase) {
	for _, phd := range phds {
		if phd.Type == RUNWAY {
			rw := &model.Runway{
				First:  phd.FirstPtIdx,
				Count:  phd.PointCount,
				Yaw:    phd.Data,
				Points: []*model.Position{},
			}
			airbase.Runways = append(airbase.Runways, rw)
		}
	}
	for _, rw := range airbase.Runways {
		for _, pdx := range pdxs {
			if pdx.Type == RUNWAY && pdx.Num >= rw.First && pdx.Num < rw.First+rw.Count {
				rw.Points = append(rw.Points, &model.Position{
					X: pdx.OffsetX,
					Y: pdx.OffsetY,
				})
			}
		}
		ComputeRunwayGeometry(rw)
	}
}

func ComputeRunwayGeometry(rw *model.Runway) bool {
	if len(rw.Points) < 4 {
		return false
	}

	var sumX, sumY float64
	for _, p := range rw.Points {
		sumX += p.X
		sumY += p.Y
	}

	l := float64(len(rw.Points))

	rw.Pos = model.Position{X: sumX / l, Y: sumY / l}

	dx := rw.Points[3].X - rw.Points[0].X
	dy := rw.Points[3].Y - rw.Points[0].Y
	rw.Length = math.Hypot(dx, dy)

	// yaw/heading: degrees(atan2(dx, dy)) (Achtung: absichtlich dx,dy in dieser Reihenfolge wie im Python)
	rw.Heading = math.Mod((math.Atan2(dx, dy)*180/math.Pi)+360, 360)

	dxW := rw.Points[1].X - rw.Points[0].X
	dyW := rw.Points[1].Y - rw.Points[0].Y
	rw.Width = math.Hypot(dxW, dyW)

	return true
}
