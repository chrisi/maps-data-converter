package model

type Station struct {
	CampId  int      `json:"campId"`
	OcdIdx  int      `json:"ocdIdx"`
	Name    string   `json:"name"`
	Country string   `json:"country"`
	Type    string   `json:"type"`
	Pos     Position `json:"pos"` // raw BMS position in feet
	Details *Details `json:"details,omitempty"`
}

type Details struct {
	Name   string      `json:"name"`
	Elev   string      `json:"elev"`
	Range  string      `json:"range"`
	Rwy    string      `json:"rwy"`
	Tcn    string      `json:"tcn"`
	Atis   string      `json:"atis"`
	Ops    string      `json:"ops"`
	Gnd    string      `json:"gnd"`
	Twr    string      `json:"twr"`
	AppDep string      `json:"appdep"`
	Charts []*ChartRef `json:"charts,omitempty"`
}

type ChartRef struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Page   *int   `json:"page,omitempty"`
	Width  *int   `json:"width,omitempty"`
	Height *int   `json:"height,omitempty"`
}
