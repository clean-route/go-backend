package utils

func GetMassFromMode(mode string) uint32 {
	switch mode {
	case "car":
		return 2500
	case "truck":
		return 1500
	case "scooter":
		return 150
	default:
		return 0
	}
}
