package utils

func GetMassFromMode(mode string) uint32 {
	switch mode {
	case "car":
		return 1800 // Average car mass in kg
	case "driving-traffic":
		return 1800 // Average car mass in kg
	case "truck":
		return 8000 // Average truck mass in kg
	case "scooter":
		return 150 // Average scooter mass in kg
	default:
		return 0
	}
}
