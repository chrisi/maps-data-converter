package model

type StationFreq struct {
	Theater  string   `json:"theater"`
	Key      string   `json:"key"`
	Name     string   `json:"name"`
	Icao     string   `json:"icao"`
	Tacan    string   `json:"tacan"`
	Band     string   `json:"band"`
	TowerUhf string   `json:"towerUhf"`
	TowerVhf string   `json:"towerVhf"`
	Runways  []string `json:"runways"`
	Ops      string   `json:"ops"`
	Ground   string   `json:"ground"`
	Approach string   `json:"approach"`
	Atis     string   `json:"atis"`
}
