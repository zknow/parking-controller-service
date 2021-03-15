package controllers

import (
	log "github.com/gogf/gf/os/glog"

	"github.com/zknow/parkingCharge/adapterClient"
	"github.com/zknow/parkingCharge/api/camera"
	"github.com/zknow/parkingCharge/api/screen"
	"github.com/zknow/parkingCharge/api/xps"
	"github.com/zknow/parkingCharge/services"

	"time"

	"github.com/spf13/viper"
)

//車辨出場事件
const (
	event_normal = iota
	event_camServe
	event_cardServe
	event_screenServe
)

// 車辨出場
type CameraOut struct {
	ScanCount          int
	ServeEvent         int
	PlatID             string
	CarPlatPicturePath string
}

type savePassInfo struct {
	ID         string
	AccessTime string
	UserType   string
	EventCode  string
	Msg        string
}

var passInfo *savePassInfo

// 車辨出場判斷程序
func (rcv *CameraOut) Serve() {
	time.Sleep(time.Second)
	switch rcv.ServeEvent {
	case event_normal:
		rcv.normal()
	case event_camServe:
		rcv.camServe()
	case event_cardServe:
		rcv.cardServe()
	case event_screenServe:
		rcv.screenServe()
	}
}
func (rcv *CameraOut) normal() {
	if rcv.ScanCount < 2 {
		if rcv.camServe() {
			rcv.ScanCount = 0
		} else {
			rcv.ScanCount++
		}
		time.Sleep(time.Second * 2)
		return
	}
	rcv.ServeEvent = event_cardServe
}

func (rcv *CameraOut) camServe() bool {
	var err error
	screen.LicenseRec()
	user := map[string]string{}
	ID, err := camera.GetInfo()
	if err != nil {
		rcv.showError("網路異常")
	}
	if ID == "" {
		return false
	}
	ID, msg, err := services.CompareCarLicence(ID)
	if err != nil {
		rcv.showError("網路異常")
		return false
	}
	if ID == "" {
		rcv.showError(msg)
		return false
	}
	user["ID"] = ID
	user["AccessTime"] = rcv.nowTime()
	user["PayID"] = user["ID"]

	verify, err := services.VerifyPassLegal(user["ID"], user["AccessTime"], "0", "0", "0")
	if err != nil {
		rcv.showError("網路異常" + err.Error())
		return false
	}
	if !verify.Legal {
		services.Block(user["ID"], verify.EventCode, verify.Msg)
		return false
	}
	user["UserType"], err = services.VerifyUserType(user["ID"])
	if err != nil {
		rcv.showError("網路異常" + err.Error())
		return false
	}
	if user["UserType"] == "Manager" {
		services.Pass(user["ID"], user["AccessTime"], user["UserType"], viper.GetString("ManaIn"), "管理卡進場")
		rcv.Reset()
		return true
	}

	if verify.Amount == "0" {
		user["PayType"] = "cash"
		xps.DbUpdate(viper.GetString("deviceno"), user["ID"], user["AccessTime"], user["PayType"],
			viper.GetString("Device.CarType"), user["PayID"], verify.Sn)
		services.Pass(user["ID"], user["AccessTime"], user["UserType"], verify.EventCode, verify.Msg)
		rcv.Reset()
		return true
	}
	rcv.PlatID = user["ID"]
	rcv.ServeEvent = event_cardServe
	return true
}

// 車辨出場車號合法程序
func (rcv *CameraOut) cardServe() {
	var user map[string]string
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
	user["PayID"] = user["ID"]
	if rcv.PlatID != "" {
		user["ID"] = rcv.PlatID
	}
	verify, err := services.VerifyPassLegal(user["ID"], user["AccessTime"], user["Heart"], user["TransferCost"], user["MRTTime"])
	if err != nil {
		rcv.showError("網路異常" + err.Error())
		return
	}
	if !verify.Legal {
		services.Block(user["ID"], verify.EventCode, verify.Msg)
		rcv.ServeEvent = event_screenServe
		screen.GetCarNumber()
		return
	}

	if verify.Amount == "0" {
		user["PayType"] = "cash"
		xps.DbUpdate(viper.GetString("deviceno"), user["ID"], user["AccessTime"], user["PayType"],
			viper.GetString("Device.CarType"), user["PayID"], verify.Sn)
		services.Pass(user["ID"], user["AccessTime"], user["UserType"], verify.EventCode, verify.Msg)
		rcv.Reset()
		return
	}
	if !services.CheckCardsBalanceEnought(user["Balance"], user["Autoload"], verify.Amount) {
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

	rcv.savePassInfo(user["ID"], user["AccessTime"], user["UserType"], verify.EventCode, verify.Msg)
	// 開始發票流程
	rcv.ServeEvent = event_screenServe
	screen.InvoiceFlow()
}

// 使用者輸入資訊服務
func (c *CameraOut) screenServe() {
	event, err := adapterClient.ListenScreenMsg()
	if err != nil {
		screen.ShowText("設備異常")
		log.Error("[Error] 與unix Server 斷線", err)
		return
	}
	if event == nil {
		return
	}
	switch event.EventName {
	case "Invoice":
		log.Info(event.Invoice.SaveInvoice)
		log.Info(event.Invoice.TxNumber)
		services.Pass(passInfo.ID, passInfo.AccessTime, passInfo.UserType, passInfo.EventCode, passInfo.Msg)
		c.Reset()
	case "CarNumber":
		log.Info(event.CarNumber)
		screen.GetCarPlat(c.CarPlatPicturePath)
	case "CarPlat":
		log.Info(event.CarPlat)
		c.PlatID = event.CarPlat.Number
		c.ServeEvent = event_cardServe
	default:
	}
}

func (rcv *CameraOut) Reset() {
	rcv = &CameraOut{ServeEvent: event_normal, PlatID: "", ScanCount: 0, CarPlatPicturePath: ""}
}

// 儲存 passinfo 給invoice流程
func (rcv *CameraOut) savePassInfo(id, accessTime, userType, eventCode, msg string) {
	passInfo = &savePassInfo{ID: id, AccessTime: accessTime, UserType: userType, EventCode: eventCode, Msg: msg}
}

func (rcv *CameraOut) nowTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (rcv *CameraOut) showError(msg string) {
	screen.ShowText(msg)
	time.Sleep(time.Second * 2)
}
