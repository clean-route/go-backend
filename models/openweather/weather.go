package openweather

// Weather represents the weather data in the JSON
type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// CurrentWeather represents the current weather data in the JSON
type CurrentWeather struct {
	Dt               int64     `json:"dt"`
	Sunrise          int64     `json:"sunrise"`
	Sunset           int64     `json:"sunset"`
	Temp             float64   `json:"temp"`
	FeelsLike        float64   `json:"feels_like"`
	Pressure         float64   `json:"pressure"`
	Humidity         float64   `json:"humidity"`
	DewPoint         float64   `json:"dew_point"`
	RelativeHumidity float64   `json:"relative_humidity"`
	Uvi              float64   `json:"uvi"`
	Clouds           float64   `json:"clouds"`
	Visibility       float64   `json:"visibility"`
	WindSpeed        float64   `json:"wind_speed"`
	WindDeg          float64   `json:"wind_deg"`
	Weather          []Weather `json:"weather"`
}

// MinutelyData represents the minutely data in the JSON
type MinutelyData struct {
	Dt            int64 `json:"dt"`
	Precipitation int   `json:"precipitation"`
}

// HourlyData represents the hourly data in the JSON
type HourlyData struct {
	Dt               int64     `json:"dt"`
	Temp             float64   `json:"temp"`
	FeelsLike        float64   `json:"feels_like"`
	Pressure         float64   `json:"pressure"`
	Humidity         float64   `json:"humidity"`
	DewPoint         float64   `json:"dew_point"`
	Uvi              float64   `json:"uvi"`
	Clouds           float64   `json:"clouds"`
	Visibility       float64   `json:"visibility"`
	WindSpeed        float64   `json:"wind_speed"`
	WindDeg          float64   `json:"wind_deg"`
	WindGust         float64   `json:"wind_gust"`
	Weather          []Weather `json:"weather"`
	Pop              float64   `json:"pop"`
	RelativeHumidity float64   `json:"relative_humidity"`
}

// WeatherData represents the overall weather data in the JSON
type WeatherData struct {
	Lat            float64        `json:"lat"`
	Lon            float64        `json:"lon"`
	Timezone       string         `json:"timezone"`
	TimezoneOffset int            `json:"timezone_offset"`
	Current        CurrentWeather `json:"current"`
	Minutely       []MinutelyData `json:"minutely"`
	Hourly         []HourlyData   `json:"hourly"`
}



