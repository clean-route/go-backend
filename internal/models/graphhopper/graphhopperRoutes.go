package graphhopperroutes

type Waypoint struct {
	Type        string        `json:"type"`
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
	TotalEnergy      float64                `json:"total_energy"`
	TotalExposure    float64                `json:"total_exposure"`
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

type RouteList struct {
	Source      []float64 `json:"source"`
	Destination []float64 `json:"destination"`
	DelayCode   uint8     `json:"delayCode"`
	Mode        string    `json:"mode"`
	RoutePref   string    `json:"route_preference"`
	Fastest     Path     `json:"fastest"`
	Shortest    Path     `json:"shortest"`
	LeapG        Path     `json:"leap_graphhopper"`
	Lco2G        Path     `json:"lco2_graphhopper"`
	Balanced    Path     `json:"balanced"`
}
