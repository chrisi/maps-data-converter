package model

import "encoding/xml"

type CampObjData struct {
	XMLName  xml.Name  `xml:"CampObjData" json:"-"`
	CampObjs []CampObj `xml:"CampObj" json:"campObjs"`
}

type CampObj struct {
	CampId int `xml:"CampId,attr" json:"campId"`

	CampName  string  `xml:"CampName" json:"campName"`
	OcdIndex  int     `xml:"OcdIndex" json:"ocdIndex"`
	Heading   float64 `xml:"Heading" json:"heading"`
	PositionX float64 `xml:"PositionX" json:"positionX"`
	PositionY float64 `xml:"PositionY" json:"positionY"`
	PositionZ float64 `xml:"PositionZ" json:"positionZ"`
}
