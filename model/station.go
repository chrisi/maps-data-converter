package model

type Station struct {
	Name    string   `json:"name"`
	Country string   `json:"country"`
	Type    string   `json:"type"`
	PosX    int      `json:"posx"`
	PosY    int      `json:"posy"`
	Details *Details `json:"details,omitempty"`
}

type Details struct {
	Name   string `json:"name"`
	Lat    string `json:"lat"`
	Long   string `json:"long"`
	Elev   string `json:"elev"`
	Rwy    string `json:"rwy"`
	Tcn    string `json:"tcn"`
	Atis   string `json:"atis"`
	Ops    string `json:"ops"`
	Gnd    string `json:"gnd"`
	Twr    string `json:"twr"`
	AppDep string `json:"appdep"`
}
