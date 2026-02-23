package model

type Airbase struct {
	Id      int
	Name    string
	Pos     Position
	OcdIdx  int
	Runways []*Runway
}

type Runway struct {
	Count   int
	First   int
	Yaw     float64
	Points  []*Position
	Pos     Position
	Length  float64
	Width   float64
	Heading float64
}

type Position struct {
	X float64
	Y float64
}
