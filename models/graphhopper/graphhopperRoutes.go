package models

type Waypoint struct {
	Type        string    `json:"type"`
	Coordinates []Coordinates `json:"coordinates"`
}

type Coordinates [3]float64

type Instruction struct {
	StreetRef  string  `json:"street_ref"`
	Distance   float64 `json:"distance"`
	Sign       int     `json:"sign"`
	Interval   []int   `json:"interval"`
	Text       string  `json:"text"`
	Time       int     `json:"time"`
	StreetName string  `json:"street_name"`
}

type Path struct {
	Distance         float64                `json:"distance"`
	Weight           float64                `json:"weight"`
	Time             int                    `json:"time"`
	Transfers        int                    `json:"transfers"`
	PointsEncoded    bool                   `json:"points_encoded"`
	BBox             []float64              `json:"bbox"`
	Points           Waypoint               `json:"points"`
	Instructions     []Instruction          `json:"instructions"`
	Legs             []interface{}          `json:"legs"`
	Details          map[string]interface{} `json:"details"`
	Ascend           float64                `json:"ascend"`
	Descend          float64                `json:"descend"`
	SnappedWaypoints Waypoint               `json:"snapped_waypoints"`
}

type Hint struct {
	VisitedNodesSum     int     `json:"visited_nodes.sum"`
	VisitedNodesAverage float64 `json:"visited_nodes.average"`
}

type Info struct {
	Copyrights []string `json:"copyrights"`
	Took       int      `json:"took"`
}

type RouteData struct {
	Hints Hint   `json:"hints"`
	Info  Info   `json:"info"`
	Paths []Path `json:"paths"`
}
