package reader

import (
	"github.com/tupyy/lazylogger/internal/conf"
)

func mapFromArray(configurations []conf.LoggerConfiguration) map[int]conf.LoggerConfiguration {
	// map the array to map. much easier to check if logger exists
	confMap := make(map[int]conf.LoggerConfiguration)
	for i := 0; i < len(configurations); i++ {
		confMap[i] = configurations[i]
	}

	return confMap

}
