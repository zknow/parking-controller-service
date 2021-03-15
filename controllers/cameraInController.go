package controllers

import (
	"time"

	log "github.com/gogf/gf/os/glog"

	"github.com/zknow/parkingCharge/api/camera"
	"github.com/zknow/parkingCharge/api/screen"
	"github.com/zknow/parkingCharge/services"

	"github.com/spf13/viper"
)

// 車辨進場
type CameraIn struct {
	ScanCount int
}

// 車辨進場流程
func (c *CameraIn) Serve() {
	if c.ScanCount < 2 {
		ok := c.camServe()
		if ok {
			c.ScanCount = 0
		} else {
			c.ScanCount++
		}
		time.Sleep(time.Second * 1)
		return
	}
	if c.cardServe() {
		c.ScanCount = 0
	}
}

// 車辨進場程序
func (c *CameraIn) camServe() bool {
	var err error
	screen.LicenseRec()
	user := map[string]string{}
	ID, err := camera.GetInfo()
	if err != nil {
		c.showError("網路異常")
	}
	if ID == "" {
		return false
	}
	ID, msg, err := services.CompareCarLicence(ID)
	if err != nil {
		c.showError("網路異常")
		return false
	}
	if ID == "" {
		c.showError(msg)
		return false
	}
	user["ID"] = ID
	user["PayID"] = user["ID"]
	user["AccessTime"] = c.nowTime()

	verify, err := services.VerifyPassLegal(user["ID"], user["AccessTime"], user["Heart"], user["TransferCost"], user["MRTTime"])
	if err != nil {
		c.showError("網路異常" + err.Error())
		return false
	}
	if !verify.Legal {
		services.Block(user["ID"], verify.EventCode, verify.Msg)
		return false
	}
	user["UserType"], err = services.VerifyUserType(user["ID"])
	if err != nil {
		c.showError("網路異常" + err.Error())
		return false
	}
	if user["UserType"] == "Manager" {
		services.Pass(user["ID"], user["AccessTime"], user["UserType"], viper.GetString("ManaIn"), "管理卡進場")
		return true
	}
	services.Pass(user["ID"], user["AccessTime"], user["UserType"], verify.EventCode, verify.Msg)
	return true
}

// 卡片進場程序
func (c *CameraIn) cardServe() bool {
	var user map[string]string
	services.InductionCardsHint(services.GetGPIOStat("FullCar"))
	info, err := services.GetCardInfo()
	if err != nil {
		log.Error("設備異常")
		return false
	}

	if info == nil {
		return false
	}

	log.Info("\n卡片資訊:", info)

	user = info

	user["AccessTime"] = c.nowTime()
	user["UserType"], err = services.VerifyUserType(user["ID"])
	if err != nil {
		c.showError("網路異常" + err.Error())
		return false
	}
	if user["UserType"] == "Manager" {
		services.Pass(user["ID"], user["AccessTime"], user["UserType"], viper.GetString("ManaIn"), "")
		return true
	}
	verify, err := services.VerifyPassLegal(user["ID"], user["AccessTime"], user["Heart"], user["TransferCost"], user["MRTTime"])
	if err != nil {
		c.showError("網路異常" + err.Error())
		return false
	}
	if !verify.Legal {
		services.Block(user["ID"], verify.EventCode, verify.Msg)
		return false
	}
	if user["UserType"] != "Monthly" && !services.IsPassMinAmountEnought(user["Balance"]) {
		services.Block(user["ID"], viper.GetString("NormalCardInFail"), "餘額不足")
		return false
	}
	services.Pass(user["ID"], user["AccessTime"], user["UserType"], verify.EventCode, verify.Msg)
	return true
}

func (c *CameraIn) Reset() {
	c = &CameraIn{}
}

func (c *CameraIn) nowTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (c *CameraIn) showError(msg string) {
	screen.ShowText(msg)
	time.Sleep(time.Second * 2)
}
