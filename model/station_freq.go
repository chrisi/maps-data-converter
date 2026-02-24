package model

type StationFreq struct {
	CampId   int      `json:"campId"`
	Tacan    string   `json:"tacan"`
	Range    int      `json:"range"`
	Band     string   `json:"band"`
	TowerUhf string   `json:"towerUhf"`
	TowerVhf string   `json:"towerVhf"`
	Runways  []string `json:"runways"`
	Ops      string   `json:"ops"`
	Ground   string   `json:"ground"`
	Approach string   `json:"approach"`
	Atis     string   `json:"atis"`
}
