package model

import "encoding/xml"

type OCDRecords struct {
	XMLName xml.Name `xml:"OCDRecords" json:"-"`
	OCDs    []OCD    `xml:"OCD" json:"ocds"`
}

type OCD struct {
	Num int `xml:"Num,attr" json:"num"`

	CtIdx         int    `xml:"CtIdx" json:"ctIdx"`
	Name          string `xml:"Name" json:"name"`
	DataRate      int    `xml:"DataRate" json:"dataRate"`
	DeaggDistance int    `xml:"DeaggDistance" json:"deaggDistance"`

	DetNoMove  float64 `xml:"Det_NoMove" json:"detNoMove"`
	DetFoot    float64 `xml:"Det_Foot" json:"detFoot"`
	DetWheeled float64 `xml:"Det_Wheeled" json:"detWheeled"`
	DetTracked float64 `xml:"Det_Tracked" json:"detTracked"`
	DetLowAir  float64 `xml:"Det_LowAir" json:"detLowAir"`
	DetAir     float64 `xml:"Det_Air" json:"detAir"`
	DetNaval   float64 `xml:"Det_Naval" json:"detNaval"`
	DetRail    float64 `xml:"Det_Rail" json:"detRail"`

	DamNone          int `xml:"Dam_None" json:"damNone"`
	DamPenetration   int `xml:"Dam_Penetration" json:"damPenetration"`
	DamHighExplosive int `xml:"Dam_HighExplosive" json:"damHighExplosive"`
	DamHeave         int `xml:"Dam_Heave" json:"damHeave"`
	DamIncendairy    int `xml:"Dam_Incendairy" json:"damIncendairy"`
	DamProximity     int `xml:"Dam_Proximity" json:"damProximity"`
	DamKinetic       int `xml:"Dam_Kinetic" json:"damKinetic"`
	DamHydrostatic   int `xml:"Dam_Hydrostatic" json:"damHydrostatic"`
	DamChemical      int `xml:"Dam_Chemical" json:"damChemical"`
	DamNuclear       int `xml:"Dam_Nuclear" json:"damNuclear"`
	DamOther         int `xml:"Dam_Other" json:"damOther"`

	ObjectiveIcon int `xml:"ObjectiveIcon" json:"objectiveIcon"`
	RadarFeature  int `xml:"RadarFeature" json:"radarFeature"`
}

type FEDRecords struct {
	XMLName xml.Name `xml:"FEDRecords" json:"-"`
	FEDs    []FED    `xml:"FED" json:"feds"`
}

type FED struct {
	Num int `xml:"Num,attr" json:"num"`

	FeatureCtIdx int     `xml:"FeatureCtIdx" json:"featureCtIdx"`
	Value        int     `xml:"Value" json:"value"`
	OffsetX      float64 `xml:"OffsetX" json:"offsetX"`
	OffsetY      float64 `xml:"OffsetY" json:"offsetY"`
	OffsetZ      float64 `xml:"OffsetZ" json:"offsetZ"`
	Heading      float64 `xml:"Heading" json:"heading"`
}

type PHDRecords struct {
	XMLName xml.Name `xml:"PHDRecords" json:"-"`
	PHDs    []PHD    `xml:"PHD" json:"phds"`
}

type PHD struct {
	Num int `xml:"Num,attr" json:"num"`

	ObjIdx         int     `xml:"ObjIdx" json:"objIdx"`
	Type           int     `xml:"Type" json:"type"`
	PointCount     int     `xml:"PointCount" json:"pointCount"`
	Data           float64 `xml:"Data" json:"data"`
	FirstPtIdx     int     `xml:"FirstPtIdx" json:"firstPtIdx"`
	RunwayTexture  int     `xml:"RunwayTexture" json:"runwayTexture"`
	RunwayNumber   int     `xml:"RunwayNumber" json:"runwayNumber"`
	LandingPattern int     `xml:"LandingPattern" json:"landingPattern"`
}

type PDRecords struct {
	XMLName xml.Name `xml:"PDRecords" json:"-"`
	PDs     []PD     `xml:"PD" json:"pds"`
}

type PD struct {
	Num int `xml:"Num,attr" json:"num"`

	OffsetX   float64 `xml:"OffsetX" json:"offsetX"`
	OffsetY   float64 `xml:"OffsetY" json:"offsetY"`
	OffsetZ   float64 `xml:"OffsetZ" json:"offsetZ"`
	MaxHeight float64 `xml:"MaxHeight" json:"maxHeight"`
	MaxWidth  float64 `xml:"MaxWidth" json:"maxWidth"`
	MaxLength float64 `xml:"MaxLength" json:"maxLength"`
	Type      int     `xml:"Type" json:"type"`
	Flags     int     `xml:"Flags" json:"flags"`
	Heading   float64 `xml:"Heading" json:"heading"`
}
