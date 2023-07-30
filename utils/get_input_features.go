package utils

import (
	"github.com/sadityakumar9211/clean-route-backend/models"
	"github.com/sadityakumar9211/clean-route-backend/models/openweather"
)

func GetInputFeatures(sourceWeather openweather.WeatherData, destinationWeather openweather.WeatherData, delayCode uint8) models.FeatureVector {
	var inputFeatures models.FeatureVector
	inputFeatures.ITEMP = (sourceWeather.Current.Temp + destinationWeather.Current.Temp) / 2
	inputFeatures.IRH = (sourceWeather.Current.RelativeHumidity + destinationWeather.Current.RelativeHumidity) / 2
	inputFeatures.IWD = (sourceWeather.Current.WindDeg + destinationWeather.Current.WindDeg) / 2
	inputFeatures.IWS = (sourceWeather.Current.WindSpeed + destinationWeather.Current.WindSpeed) / 2
	switch delayCode {
	case 0:
		// 30 min delay
		// take the average between the curr and
		inputFeatures.FTEMP = (inputFeatures.ITEMP + sourceWeather.Hourly[1].Temp + destinationWeather.Hourly[1].Temp) / 3
		inputFeatures.FRH = (inputFeatures.IRH + GetRelativeHumidity(sourceWeather.Hourly[1].DewPoint, sourceWeather.Hourly[1].Temp) + GetRelativeHumidity(destinationWeather.Hourly[1].DewPoint, destinationWeather.Hourly[1].Temp)) / 3
		inputFeatures.FWD = (inputFeatures.IWD + sourceWeather.Hourly[1].WindDeg + destinationWeather.Hourly[1].WindDeg) / 3
		inputFeatures.FWS = (inputFeatures.IWS + sourceWeather.Hourly[1].WindSpeed + sourceWeather.Hourly[1].WindSpeed) / 3
	case 1:
		// 60 min delay
		inputFeatures.FTEMP = (sourceWeather.Hourly[1].Temp + destinationWeather.Hourly[1].Temp) / 2
		inputFeatures.FRH = (GetRelativeHumidity(sourceWeather.Hourly[1].DewPoint, sourceWeather.Hourly[1].Temp) + GetRelativeHumidity(destinationWeather.Hourly[1].DewPoint, destinationWeather.Hourly[1].Temp)) / 2
		inputFeatures.FWD = (sourceWeather.Hourly[1].WindDeg + destinationWeather.Hourly[1].WindDeg) / 2
		inputFeatures.FWS = (sourceWeather.Hourly[1].WindSpeed + destinationWeather.Hourly[1].WindSpeed) / 2

	case 2:
		// 2 hr delay
		inputFeatures.FTEMP = (sourceWeather.Hourly[2].Temp + destinationWeather.Hourly[2].Temp) / 2
		inputFeatures.FRH = (GetRelativeHumidity(sourceWeather.Hourly[2].DewPoint, sourceWeather.Hourly[2].Temp) + GetRelativeHumidity(destinationWeather.Hourly[2].DewPoint, destinationWeather.Hourly[2].Temp)) / 2
		inputFeatures.FWD = (sourceWeather.Hourly[2].WindDeg + destinationWeather.Hourly[2].WindDeg) / 2
		inputFeatures.FWS = (sourceWeather.Hourly[2].WindSpeed + destinationWeather.Hourly[2].WindSpeed) / 2
	case 3:
		// 3 hr delay
		inputFeatures.FTEMP = (sourceWeather.Hourly[3].Temp + destinationWeather.Hourly[3].Temp) / 2
		inputFeatures.FRH = (GetRelativeHumidity(sourceWeather.Hourly[3].DewPoint, sourceWeather.Hourly[3].Temp) + GetRelativeHumidity(destinationWeather.Hourly[3].DewPoint, destinationWeather.Hourly[3].Temp)) / 2
		inputFeatures.FWD = (sourceWeather.Hourly[3].WindDeg + destinationWeather.Hourly[3].WindDeg) / 2
		inputFeatures.FWS = (sourceWeather.Hourly[3].WindSpeed + destinationWeather.Hourly[3].WindSpeed) / 2
	case 4:
		// 4 hr delay
		inputFeatures.FTEMP = (sourceWeather.Hourly[4].Temp + destinationWeather.Hourly[4].Temp) / 2
		inputFeatures.FRH = (GetRelativeHumidity(sourceWeather.Hourly[4].DewPoint, sourceWeather.Hourly[4].Temp) + GetRelativeHumidity(destinationWeather.Hourly[4].DewPoint, destinationWeather.Hourly[4].Temp)) / 2
		inputFeatures.FWD = (sourceWeather.Hourly[4].WindDeg + destinationWeather.Hourly[4].WindDeg) / 2
		inputFeatures.FWS = (sourceWeather.Hourly[4].WindSpeed + destinationWeather.Hourly[4].WindSpeed) / 2
	case 5:
		// 5 hr delay
		inputFeatures.FTEMP = (sourceWeather.Hourly[5].Temp + destinationWeather.Hourly[5].Temp) / 2
		inputFeatures.FRH = (GetRelativeHumidity(sourceWeather.Hourly[5].DewPoint, sourceWeather.Hourly[5].Temp) + GetRelativeHumidity(destinationWeather.Hourly[5].DewPoint, destinationWeather.Hourly[5].Temp)) / 2
		inputFeatures.FWD = (sourceWeather.Hourly[5].WindDeg + destinationWeather.Hourly[5].WindDeg) / 2
		inputFeatures.FWS = (sourceWeather.Hourly[5].WindSpeed + destinationWeather.Hourly[5].WindSpeed) / 2
	case 6:
		// 6 hr delay
		inputFeatures.FTEMP = (sourceWeather.Hourly[6].Temp + destinationWeather.Hourly[6].Temp) / 2
		inputFeatures.FRH = (GetRelativeHumidity(sourceWeather.Hourly[6].DewPoint, sourceWeather.Hourly[6].Temp) + GetRelativeHumidity(destinationWeather.Hourly[6].DewPoint, destinationWeather.Hourly[6].Temp)) / 2
		inputFeatures.FWD = (sourceWeather.Hourly[6].WindDeg + destinationWeather.Hourly[6].WindDeg) / 2
		inputFeatures.FWS = (sourceWeather.Hourly[6].WindSpeed + destinationWeather.Hourly[6].WindSpeed) / 2
	default:
		// same as current values
		inputFeatures.FTEMP = (sourceWeather.Current.Temp + destinationWeather.Current.Temp) / 2
		inputFeatures.FRH = (sourceWeather.Current.RelativeHumidity + destinationWeather.Current.RelativeHumidity) / 2
		inputFeatures.FWD = (sourceWeather.Current.WindDeg + destinationWeather.Current.WindDeg) / 2
		inputFeatures.FWS = (sourceWeather.Current.WindSpeed + destinationWeather.Current.WindSpeed) / 2

	}
	return inputFeatures
}
