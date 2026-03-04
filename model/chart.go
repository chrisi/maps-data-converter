package model

type Chart struct {
	Key     string `json:"key"`
	Theater string `json:"theater"`
	Code    string `json:"code"`
	Country string `json:"country"`
	Station string `json:"station"`
	Icao    string `json:"icao"`
	Path    string `json:"path"`
	File    string `json:"file"`
	Name    string `json:"name"`
}
