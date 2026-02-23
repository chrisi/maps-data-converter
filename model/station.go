package model

type Station struct {
	OcdIdx  int      `json:"ocdIdx"`
	Name    string   `json:"name"`
	Country string   `json:"country"`
	Type    string   `json:"type"`
	PosX    int      `json:"posx"`
	PosY    int      `json:"posy"`
	Details *Details `json:"details,omitempty"`
}

type Details struct {
	Name   string   `json:"name"`
	Elev   string   `json:"elev"`
	Rwy    string   `json:"rwy"`
	Tcn    string   `json:"tcn"`
	Atis   string   `json:"atis"`
	Ops    string   `json:"ops"`
	Gnd    string   `json:"gnd"`
	Twr    string   `json:"twr"`
	AppDep string   `json:"appdep"`
	Charts []*Chart `json:"charts,omitempty"`
}

type Chart struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Page   *int   `json:"page,omitempty"`
	Width  *int   `json:"width,omitempty"`
	Height *int   `json:"height,omitempty"`
}
