package models

// Define structs to represent the JSON data

type Waypoint struct {
	Distance float64   `json:"distance"`
	Name     string    `json:"name"`
	Location []float64 `json:"location"`
}

type Step struct {
	Intersections   []Intersection `json:"intersections"`
	Maneuver        Maneuver       `json:"maneuver"`
	Name            string         `json:"name"`
	WeightTypical   float64        `json:"weight_typical"`
	DurationTypical float64        `json:"duration_typical"`
	Duration        float64        `json:"duration"`
	Distance        float64        `json:"distance"`
	DrivingSide     string         `json:"driving_side"`
	Weight          float64        `json:"weight"`
	Mode            string         `json:"mode"`
	Geometry        Geometry       `json:"geometry"`
}

type Intersection struct {
	Bearings        []int           `json:"bearings"`
	Entry           []bool          `json:"entry"`
	MapboxStreetsV8 MapboxStreetsV8 `json:"mapbox_streets_v8"`
	IsUrban         bool            `json:"is_urban"`
	AdminIndex      int             `json:"admin_index"`
	Out             int             `json:"out"`
	GeometryIndex   int             `json:"geometry_index"`
	Location        []float64       `json:"location"`
}

type MapboxStreetsV8 struct {
	Class string `json:"class"`
}

type Maneuver struct {
	Type          string    `json:"type"`
	Instruction   string    `json:"instruction"`
	BearingAfter  int       `json:"bearing_after"`
	BearingBefore int       `json:"bearing_before"`
	Location      []float64 `json:"location"`
}

type Geometry struct {
	Coordinates [][]float64 `json:"coordinates"`
	Type        string      `json:"type"`
}

type Leg struct {
	ViaWaypoints    []interface{} `json:"via_waypoints"`
	Admins          []Admin       `json:"admins"`
	WeightTypical   float64       `json:"weight_typical"`
	DurationTypical float64       `json:"duration_typical"`
	Weight          float64       `json:"weight"`
	Duration        float64       `json:"duration"`
	Steps           []Step        `json:"steps"`
	Distance        float64       `json:"distance"`
	Summary         string        `json:"summary"`
}

type Admin struct {
	Iso31661Alpha3 string `json:"iso_3166_1_alpha3"`
	Iso31661       string `json:"iso_3166_1"`
}

type Route struct {
	WeightTypical   float64    `json:"weight_typical"`
	Waypoints       []Waypoint `json:"waypoints"`
	DurationTypical float64    `json:"duration_typical"`
	WeightName      string     `json:"weight_name"`
	Weight          float64    `json:"weight"`
	Duration        float64    `json:"duration"`
	Distance        float64    `json:"distance"`
	Legs            []Leg      `json:"legs"`
	Geometry        Geometry   `json:"geometry"`
}

type Routes struct {
	Routes []Route `json:"routes"`
	Code   string  `json:"code"`
	UUID   string  `json:"uuid"`
}
