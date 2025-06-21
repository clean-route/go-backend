package utils

import (
	"os"
	"strconv"

	graphhopper "github.com/clean-route/go-backend/internal/models/graphhopper"
)

const (
	acceleration_of_gravity = 9.8
	// Rolling resistance coefficient (typical for cars)
	rolling_resistance_coefficient = 0.01
	// Air density at sea level (kg/m³)
	air_density = 1.225
	// Drag coefficient (typical for cars)
	drag_coefficient = 0.3
	// Frontal area (m²) - typical for cars
	frontal_area = 2.0
)

func CalculateRouteEnergy(route graphhopper.Path, mode string, vehicleMass int, condition string, engineType string) float64 {
	// Use provided vehicle mass if available, otherwise fall back to default
	mass := uint32(vehicleMass)
	if mass == 0 {
		mass = GetMassFromMode(mode)
	}

	segments := route.Instructions

	var totalEnergy float64 // in Joules

	for j := 0; j < len(segments); j++ {
		startIndex := segments[j].Interval[0]
		endIndex := segments[j].Interval[1]

		heightGain := route.Points.Coordinates[endIndex][2] - route.Points.Coordinates[startIndex][2]

		distance := segments[j].Distance                  // in meters
		time := float64(segments[j].Time) / float64(1000) // now its in seconds

		if time == 0 && distance == 0 {
			continue
		}

		averageVelocity := distance / time // m/s

		// 1. Potential Energy (climbing/descending)
		potentialEnergy := float64(mass) * acceleration_of_gravity * heightGain

		// 2. Rolling Resistance Energy
		rollingResistanceEnergy := rolling_resistance_coefficient * float64(mass) * acceleration_of_gravity * distance

		// 3. Air Resistance Energy
		airResistanceEnergy := 0.5 * air_density * drag_coefficient * frontal_area * averageVelocity * averageVelocity * distance

		// 4. Kinetic Energy (acceleration/deceleration) - simplified
		// Assume average acceleration/deceleration pattern
		kineticEnergy := 0.5 * float64(mass) * averageVelocity * averageVelocity

		// Total mechanical energy for this segment
		segmentEnergy := potentialEnergy + rollingResistanceEnergy + airResistanceEnergy + kineticEnergy

		// Convert to fuel energy (accounting for engine efficiency)
		engineEfficiency := getEngineEfficiency(engineType, condition)
		fuelEnergy := segmentEnergy / engineEfficiency

		totalEnergy += fuelEnergy
	}

	// Convert from Joules to kilojoules
	energyKJ := totalEnergy / 1000

	// Return energy in kJ (not CO2 emissions)
	return energyKJ
}

// getEngineEfficiency returns engine efficiency based on type and condition
func getEngineEfficiency(engineType string, condition string) float64 {
	// Base engine efficiencies
	engineEfficiencies := map[string]float64{
		"petrol": 0.25, // 25% efficiency for petrol engines
		"diesel": 0.30, // 30% efficiency for diesel engines
		"cng":    0.28, // 28% efficiency for CNG engines
		"ev":     0.85, // 85% efficiency for electric vehicles
	}

	// Condition factors (worse condition = lower efficiency)
	conditionFactors := map[string]float64{
		"new":     1.0,  // 100% efficiency
		"good":    0.95, // 95% efficiency
		"average": 0.90, // 90% efficiency
		"okay":    0.80, // 80% efficiency
	}

	engineEfficiency := engineEfficiencies[engineType]
	if engineEfficiency == 0 && engineType != "ev" {
		engineEfficiency = engineEfficiencies["petrol"] // Default to petrol
	}

	conditionFactor := conditionFactors[condition]
	if conditionFactor == 0 {
		conditionFactor = conditionFactors["average"] // Default to average
	}

	return engineEfficiency * conditionFactor
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
