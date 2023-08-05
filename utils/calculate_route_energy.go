package utils

import (
	graphhopper "github.com/sadityakumar9211/clean-route-backend/models/graphhopper"
)

func CalculateRouteEnergy(route graphhopper.Path, mode string) float64 {

	// for carbon emssions we can find the total energy consumed by the vehicle
	// and then also consider the fuel efficiency of the vehicle. We also need to consider
	// the mass of the vehicle --> can take thvscode-file://vscode-app/Applications/Visual%20Studio%20Code%20-%20Insiders.app/Contents/Resources/app/out/vs/code/electron-sandbox/workbench/workbench.htmle average mass of the vehicle.

	mass := GetMassFromMode(mode)
	g := 9.8

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
		totalPotentialEnergy := float64(mass) * g * heightGain

		// Kinetic Energy
		totalKineticEnergy := 0.5 * float64(mass) * averageVelocity * averageVelocity

		// Total Energy = Potential + Kinetic
		totalEnergy += totalPotentialEnergy + totalKineticEnergy
	}
	// fmt.Println("Total Energy: ", totalEnergy / 1000)
	return totalEnergy / 1000
}
