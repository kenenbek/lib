package lib

import ("fmt"
)

func SIM_links_observer() string {
	var logTraffic string
	c := 0.064
	for link := range env.routesMap {
		queueLength := len(env.routesMap[link].queue)
		traffic := c * float64(queueLength)
		logTraffic += fmt.Sprintf("%.2f,", traffic)
	}
	return logTraffic
}


func SIM_storages_observer() string {
	var storageLog string
	MEGA := 1000000.
	for storage := range env.storagesMap {
		usedSize := env.storagesMap[storage].usedSize
		storageLog += fmt.Sprintf("%.4f,", float64(usedSize) / MEGA)
	}

	return storageLog
}