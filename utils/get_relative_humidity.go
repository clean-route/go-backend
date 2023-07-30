package utils

import "math"

func GetRelativeHumidity(dewPoint float64, temp float64) float64 {
	return 100 * ((math.Pow(math.E, (17.625*dewPoint)/(243.04+dewPoint))) / (math.Pow(math.E, 17.625*temp/(243.04+temp))))
}
