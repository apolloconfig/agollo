package env

import "strings"

func transformEnv(envName string) Env {
	if envName == "" {
		return UNKNOWN
	}

	upper := strings.ToUpper(envName)

	switch upper {
	case "LPT":
		return LPT
	case "FAT":
		return FAT
	case "FWS":
		return FWS
	case "UAT":
		return UAT
	case "PRO":
		return PRO
	case "PROD": //just in case
		return PRO
	case "DEV":
		return DEV
	case "LOCAL":
		return LOCAL
	case "TOOLS":
		return TOOLS
	default:
		return UNKNOWN
	}

	return UNKNOWN
}
