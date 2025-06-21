package models

// RouteRequest represents the request for route planning
type RouteRequest struct {
	Source          [2]float64 `json:"source" binding:"required"`
	Destination     [2]float64 `json:"destination" binding:"required"`
	DelayCode       uint8      `json:"delayCode"`
	Mode            string     `json:"mode" binding:"required"`
	RoutePreference string     `json:"route_preference,omitempty"`
	VehicleMass     int        `json:"vehicle_mass"`
	Condition       string     `json:"condition"`
	EngineType      string     `json:"engine_type"`
}

// PM25PredictionRequest represents the request for PM2.5 prediction
type PM25PredictionRequest struct {
	Features []FeatureVector `json:"features" binding:"required"`
}

// FeatureVector represents the feature vector for ML prediction
type FeatureVector struct {
	ITEMP     float64 `json:"ITEMP"`
	IRH       float64 `json:"IRH"`
	IWD       float64 `json:"IWD"`
	IWS       float64 `json:"IWS"`
	IPM       float64 `json:"IPM"`
	FTEMP     float64 `json:"FTEMP"`
	FRH       float64 `json:"FRH"`
	FWD       float64 `json:"FWD"`
	FWS       float64 `json:"FWS"`
	DelayCode uint8   `json:"delayCode"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
