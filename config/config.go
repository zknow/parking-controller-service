package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	log "github.com/gogf/gf/os/glog"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var confPath = "./config/config.yml"

func creatFile(t string, data interface{}) {
	var buf []byte
	var err error
	switch t {
	case "json":
		buf, err = json.MarshalIndent(data, "", "\t")
		if err != nil {
			log.Error(err)
		}
	case "yaml":
		buf, err = yaml.Marshal(data)
		if err != nil {
			log.Error(err)
		}
	}
	err = ioutil.WriteFile(confPath, buf, os.ModePerm)
	if err != nil {
		log.Fatal("config 初始化/更新失敗")
	}
	log.Info("config 初始化/更新成功")
}

func LoadConfig() {
	size, err := ioutil.ReadFile(confPath)
	if len(size) < 10 || err != nil {
		ioutil.WriteFile(confPath, []byte(defultConf), os.ModePerm)
		log.Warning("初始化config預設值,請重新設定", confPath)
	}

	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")

	if err = viper.ReadInConfig(); err != nil {
		log.Fatal("config file load 失敗:", err)
	}
}

func watchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Info("Config file changed:", e.Name)
	})
}

var defultConf = `#場佔設定
stage:
  serverIp: example
  readerEnable: "0"
  gpioEnable: "0"
  logMaxSize: "10"
  

#資料庫設定
db:
  dataBase: example
  host: example
  passWord: example
  user: example

#票證
cardsEnable:
  ecc:        1
  ipass:      0
  icash:      0
  happycash:  0

#com port
comPort:
  PrinterA: "/dev/ttymxc3"
  PrinterB: "/dev/ttymxc2"
  Reader: "/dev/"
  PLC: "/dev/"

#事件代碼(無須更動)
eventCode:
  DeductFail: A014
  DeductSuccess: A013
  ManaIn: A003
  ManaOut: A004
  PamPay: A025
`
