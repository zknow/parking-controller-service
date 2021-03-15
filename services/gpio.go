package services

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	log "github.com/gogf/gf/os/glog"

	"github.com/zknow/parkingCharge/utils"
)

type gpioStorage struct {
	ACoil   int
	BCoil   int
	Token   int
	FullCar int
	Light   int
	Gate    int
}

var gpio *gpioStorage

func init() {
	gpio = &gpioStorage{}
	filename := "gpioctl"
	if !utils.CheckFileIsExist(filename) {
		log.Fatal("gpioctl不存在")
	}
	filetime := time.Now()
	go func() {
		for {
			stat, _ := os.Stat(filename)
			if stat.ModTime() != filetime {
				b, _ := ioutil.ReadFile(filename)
				_ = json.Unmarshal(b, gpio)
			}
		}
	}()
}

func GetGPIOHD() *gpioStorage {
	return gpio
}

func GetGPIOStat(pin string) int {
	switch pin {
	case "ACoil":
		return gpio.ACoil
	case "BCoil":
		return gpio.BCoil
	case "Token":
		return gpio.Token
	case "FullCar":
		return gpio.FullCar
	case "Light":
		return gpio.Light
	default:
		return -1
	}
}
