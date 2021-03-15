package controllers

import (
	"time"

	log "github.com/gogf/gf/os/glog"

	"github.com/zknow/parkingCharge/api/screen"
	"github.com/zknow/parkingCharge/api/xps"
	"github.com/zknow/parkingCharge/services"

	"github.com/spf13/viper"
)

// 一次收費
type CardsOnce struct{}

// 一次收費程序
func (rcv *CardsOnce) Serve() {
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

	user["PayID"] = user["ID"]
	user["AccessTime"] = rcv.nowTime()
	user["UserType"], err = services.VerifyUserType(user["ID"])
	if err != nil {
		rcv.showError("網路異常" + err.Error())
		return
	}
	if user["UserType"] == "Manager" {
		services.Pass(user["ID"], user["AccessTime"], user["UserType"], viper.GetString("ManaIn"), "管理卡進場")
		return
	}
	verify, err := services.VerifyPassLegal(user["ID"], user["AccessTime"], user["Heart"], user["TransferCost"], user["MRTTime"])
	if err != nil {
		rcv.showError("網路異常" + err.Error())
		return
	}
	if !verify.Legal {
		xps.DbUpdate(viper.GetString("deviceno"), user["ID"], user["AccessTime"], user["PayType"],
			viper.GetString("Device.CarType"), user["PayID"], verify.Sn)
		services.Block(user["ID"], verify.EventCode, verify.Msg)
		return
	}
	if verify.EventCode == viper.GetString("eventCode.PamPay") {
		services.Pass(user["ID"], user["AccessTime"], user["UserType"], verify.EventCode, verify.Msg)
		return
	}

	if verify.Amount == "0" {
		user["PayType"] = "cash"
		xps.DbUpdate(viper.GetString("deviceno"), user["ID"], user["AccessTime"], user["PayType"],
			viper.GetString("Device.CarType"), user["PayID"], verify.Sn)
		services.Pass(user["ID"], user["AccessTime"], user["UserType"], verify.EventCode, verify.Msg)
		return
	}

	if !services.CheckCardsBalanceEnought(user["Balance"], user["Autoload"], verify.Amount) {
		user["PayType"] = "payFail"
		xps.DbUpdate(viper.GetString("deviceno"), user["ID"], user["AccessTime"], user["PayType"],
			viper.GetString("Device.CarType"), user["PayID"], verify.Sn)
		services.Block(user["ID"], viper.GetString("NormalCardInFail"), "餘額不足")
		return
	}
	ok, _ := services.CardsDeduct(verify.Amount)
	if !ok {
		screen.ShowText("扣款失敗")
		user["PayType"] = "payFail"
		xps.DbUpdate(viper.GetString("deviceno"), user["ID"], user["AccessTime"], user["PayType"],
			viper.GetString("Device.CarType"), user["PayID"], verify.Sn)
		services.Block(user["ID"], viper.GetString("eventCode.DeductFail"), "扣款失敗")
		return
	}
	//更新進出場紀錄
	xps.DbUpdate(viper.GetString("deviceno"), user["ID"], user["AccessTime"], user["PayType"],
		viper.GetString("Device.CarType"), user["PayID"], verify.Sn)

	//開始發票流程
	screen.InvoiceFlow()
	services.Pass(user["ID"], user["AccessTime"], user["UserType"], verify.EventCode, verify.Msg)
}

func (rcv *CardsOnce) Reset() {
	rcv = &CardsOnce{}
}

func (rcv *CardsOnce) showError(msg string) {
	screen.ShowText(msg)
	time.Sleep(time.Second * 2)
}

func (rcv *CardsOnce) nowTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
