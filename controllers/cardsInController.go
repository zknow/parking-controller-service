package controllers

import (
	"time"

	log "github.com/gogf/gf/os/glog"

	"github.com/zknow/parkingCharge/api/screen"
	"github.com/zknow/parkingCharge/api/xps"
	"github.com/zknow/parkingCharge/services"

	"github.com/spf13/viper"
)

// 卡片進場
type CardsIn struct{}

// 卡片進場程序
func (rcv *CardsIn) Serve() {
	var user map[string]string
	var err error
	services.InductionCardsHint(services.GetGPIOStat("FullCar"))
	info, err := services.GetCardInfo()
	if err != nil {
		log.Error("設備異常")
	}
	if info == nil {
		return
	}
	log.Info("\n卡片資訊:", info)
	user = info

	user["AccessTime"] = rcv.nowTime()
	user["UserType"], err = services.VerifyUserType(user["ID"])
	if err != nil {
		rcv.showError("網路異常" + err.Error())
		return
	}
	if user["UserType"] == "Manager" {
		services.Pass(user["ID"], user["AccessTime"], user["UserType"], viper.GetString("ManaIn"), "")
		return
	}
	verify, err := services.VerifyPassLegal(user["ID"], user["AccessTime"], user["Heart"], user["TransferCost"], user["MRTTime"])
	if err != nil {
		rcv.showError("網路異常" + err.Error())
		return
	}
	if !verify.Legal {
		services.Block(user["ID"], verify.EventCode, verify.Msg)
		return
	}
	if user["UserType"] != "Monthly" && !services.IsPassMinAmountEnought(user["Balance"]) {
		services.Block(user["ID"], viper.GetString("NormalCardInFail"), "餘額不足")
		return
	}
	xps.DbUpdate(viper.GetString("deviceno"), user["ID"], user["AccessTime"], user["PayType"],
		viper.GetString("Device.CarType"), user["PayID"], verify.Sn)
	services.Pass(user["ID"], user["AccessTime"], user["UserType"], verify.EventCode, verify.Msg)
}

func (rcv *CardsIn) Reset() {
	rcv = &CardsIn{}
}

func (rcv *CardsIn) showError(msg string) {
	screen.ShowText(msg)
	time.Sleep(time.Second * 2)
}

func (rcv *CardsIn) nowTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
