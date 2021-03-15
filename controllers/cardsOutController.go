package controllers

import (
	"time"

	log "github.com/gogf/gf/os/glog"

	"github.com/zknow/parkingCharge/adapterClient"
	"github.com/zknow/parkingCharge/api/screen"
	"github.com/zknow/parkingCharge/api/xps"
	"github.com/zknow/parkingCharge/services"

	"github.com/spf13/viper"
)

// 卡片出場
type CardsOut struct{}

// 卡片出場程序
func (rcv *CardsOut) Serve() {
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

	log.Info("USER:", user)

	user["AccessTime"] = rcv.nowTime()
	user["PayID"] = user["ID"]
	user["UserType"], err = services.VerifyUserType(user["ID"])

	if err != nil {
		rcv.showError("網路異常" + err.Error())
		return
	}

	if user["UserType"] == "Manager" {
		services.Pass(user["ID"], user["AccessTime"], user["UserType"], viper.GetString("ManaOut"), "")
		return
	}

	log.Println("USER:", user)
	verify, err := services.VerifyPassLegal(user["ID"], user["AccessTime"], user["Heart"], user["TransferCost"], user["MRTTime"])

	if err != nil {
		rcv.showError("網路異常" + err.Error())
		return
	}

	if !verify.Legal {
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

	if !services.CheckCardsBalanceEnought(user["Balance"], user["AutoLoad"], verify.Amount) {
		services.Block(user["ID"], viper.GetString("NormalCardInFail"), "餘額不足")
		return
	}

	if ok, _ := services.CardsDeduct(verify.Amount); !ok {
		screen.ShowText("扣款失敗")
		services.Block(user["ID"], viper.GetString("eventCode.DeductFail"), "扣款失敗")
		return
	}
	xps.DbUpdate(viper.GetString("deviceno"), user["ID"], user["AccessTime"], user["PayType"],
		viper.GetString("Device.CarType"), user["PayID"], verify.Sn)

	// 開始發票流程
	screen.InvoiceFlow()
	event, err := adapterClient.ListenScreenMsg()
	if err != nil {
		screen.ShowText("設備異常")
		log.Fatal("與unix Server 斷線", err)
		return
	}
	if event == nil {
		return
	}
	switch event.EventName {
	case "Invoice":
		log.Println(event.Invoice.SaveInvoice)
		log.Println(event.Invoice.TxNumber)
	default:
	}
	services.Pass(user["ID"], user["AccessTime"], user["UserType"], verify.EventCode, verify.Msg)
}

func (rcv *CardsOut) Reset() {
	rcv = &CardsOut{}
}

func (rcv *CardsOut) showError(msg string) {
	screen.ShowText(msg)
	time.Sleep(time.Second * 2)
}

func (rcv *CardsOut) nowTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
