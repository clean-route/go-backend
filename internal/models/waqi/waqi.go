package models

type Attribution struct {
	URL  string `json:"url"`
	Name string `json:"name"`
	Logo string `json:"logo,omitempty"`
}

type City struct {
	Geo      []float64 `json:"geo"`
	Name     string    `json:"name"`
	URL      string    `json:"url"`
	Location string    `json:"location"`
}

type Time struct {
	S   string `json:"s"`
	Tz  string `json:"tz"`
	V   int64  `json:"v"`
	Iso string `json:"iso"`
}

type DailyData struct {
	Avg int `json:"avg"`
	Day string `json:"day"`
	Max int `json:"max"`
	Min int `json:"min"`
}

type Forecast struct {
	Daily struct {
		O3   []DailyData `json:"o3"`
		PM10 []DailyData `json:"pm10"`
		PM25 []DailyData `json:"pm25"`
	} `json:"daily"`
}

type AQIData struct {
	AQI          int          `json:"aqi"`
	IDX          int          `json:"idx"`
	Attributions []Attribution `json:"attributions"`
	City         City         `json:"city"`
	Dominentpol  string       `json:"dominentpol"`
	IAQI         map[string]struct {
		V float64 `json:"v"`
	} `json:"iaqi"`
	Time     Time     `json:"time"`
	Forecast Forecast `json:"forecast"`
}

type APIResponse struct {
	Status string  `json:"status"`
	Data   AQIData `json:"data"`
}


