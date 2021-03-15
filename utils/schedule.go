package utils

import (
	log "github.com/gogf/gf/os/glog"
	"github.com/robfig/cron"
)

// 排程工具
func MakeSchedule(do func()) {
	var c *cron.Cron
	defer func() {
		if r := recover(); r != nil {
			log.Error()
		}
	}()
	c = cron.New()
	defer c.Stop()
	// spec := "*/5 * * * * ?"
	spec := "@daily"
	err := c.AddFunc(spec, func() {
		log.Info("Schedule Running...")
		do()
	})
	if err != nil {
		log.Fatal("Schedule Error : " + err.Error())
	}
	c.Start()
	select {}
}
