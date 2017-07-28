package agollo

import "strings"

func transformEnv(envName string) env {
	if envName == "" {
		return unknown
	}

	upper := strings.ToUpper(envName)

	switch upper {
	case "LPT":
		return lpt
	case "FAT":
		return fat
	case "FWS":
		return fws
	case "UAT":
		return uat
	case "PRO":
		return pro
	case "PROD": //just in case
		return pro
	case "DEV":
		return dev
	case "LOCAL":
		return local
	case "TOOLS":
		return tools
	default:
		return unknown
	}

	return unknown
}
