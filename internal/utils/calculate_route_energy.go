package utils

import (
	"os"
	"strconv"

	graphhopper "github.com/clean-route/go-backend/internal/models/graphhopper"
)

const (
	acceleration_of_gravity = 9.8
)

func CalculateRouteEnergy(route graphhopper.Path, mode string, vehicleMass int, condition string, engineType string) float64 {

	// for carbon emssions we can find the total energy consumed by the vehicle
	// and then also consider the fuel efficiency of the vehicle. We also need to consider
	// the mass of the vehicle --> can take the average mass of the vehicle.

	// Use provided vehicle mass if available, otherwise fall back to default
	mass := uint32(vehicleMass)
	if mass == 0 {
		mass = GetMassFromMode(mode)
	}

	segments := route.Instructions

	var totalEnergy float64

	for j := 0; j < len(segments); j++ {
		startIndex := segments[j].Interval[0]
		endIndex := segments[j].Interval[1]

		heightGain := route.Points.Coordinates[endIndex][2] - route.Points.Coordinates[startIndex][2]

		distance := segments[j].Distance                  // in meters
		time := float64(segments[j].Time) / float64(1000) // now its in seconds

		if time == 0 && distance == 0 {
			continue
		}
		averageVelocity := distance / time

		// Potential Energy
		totalPotentialEnergy := float64(mass) * acceleration_of_gravity * heightGain

		// Kinetic Energy
		totalKineticEnergy := 0.5 * float64(mass) * averageVelocity * averageVelocity

		// Total Energy = Potential + Kinetic
		totalEnergy += totalPotentialEnergy + totalKineticEnergy
	}

	// Apply vehicle condition and engine type factors
	emissionFactor := getEmissionFactor(condition, engineType)

	// fmt.Println("Total Energy: ", totalEnergy / 1000)
	return (totalEnergy / 1000) * emissionFactor
}

// getEmissionFactor calculates emission factor based on vehicle condition and engine type
func getEmissionFactor(condition string, engineType string) float64 {
	// Get engine type factors from environment variables with fallback defaults
	engineFactors := map[string]float64{
		"petrol": getEnvFloat("EMISSION_FACTOR_PETROL", 0.069),
		"diesel": getEnvFloat("EMISSION_FACTOR_DIESEL", 0.074),
		"cng":    getEnvFloat("EMISSION_FACTOR_CNG", 0.056),
		"ev":     getEnvFloat("EMISSION_FACTOR_EV", 0.0),
	}

	// Get condition factors from environment variables with fallback defaults
	conditionFactors := map[string]float64{
		"new":     getEnvFloat("CONDITION_FACTOR_NEW", 1.0),
		"good":    getEnvFloat("CONDITION_FACTOR_GOOD", 1.1),
		"average": getEnvFloat("CONDITION_FACTOR_AVERAGE", 1.25),
		"okay":    getEnvFloat("CONDITION_FACTOR_OKAY", 1.5),
	}

	engineFactor := engineFactors[engineType]
	if engineFactor == 0 && engineType != "ev" {
		engineFactor = engineFactors["petrol"] // Default to petrol if unknown
	}

	conditionFactor := conditionFactors[condition]
	if conditionFactor == 0 {
		conditionFactor = conditionFactors["average"] // Default to average if unknown
	}

	return engineFactor * conditionFactor
}

// getEnvFloat gets a float value from environment variable with fallback
func getEnvFloat(key string, fallback float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return fallback
}
