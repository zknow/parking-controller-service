package utils

import (
	"time"

	log "github.com/gogf/gf/os/glog"
)

func Ticker(do func(), t time.Duration, doFirst bool) {
	if doFirst {
		do()
	}
	ticker := time.NewTicker(t)
	defer ticker.Stop()
	for t := range ticker.C {
		log.Debug(t)
		do()
	}
}
