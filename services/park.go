package services

import (
	"strings"
	"time"

	"github.com/zknow/parkingCharge/api/monitor"
	"github.com/zknow/parkingCharge/api/screen"
	"github.com/zknow/parkingCharge/api/xps"

	"github.com/spf13/viper"
)

// 初始化服務
func CheckParkMode() string {
	viper.Set("Device.PassMode", "1")
	viper.Set("Camera.Enable", "0")
	passMode := viper.GetString("Device.PassMode")
	camera := viper.GetString("Camera.Enable")
	switch passMode {
	case passModeIn:
		if camera == "1" {
			return "cameraIn"
		}
		return "cardsIn"

	case passModeOut:
		if camera == "1" {
			return "cameraOut"
		}
		return "cardsOut"
	case passModeOnceCharge:
		return "cardsOnce"
	default:
		return "Error"
	}
}

// 合法開門程序
func Pass(ID, accessTime, userType, eventCode, eventMsg string) {
	identity := map[string]string{"Manager": "管理者", "Monthly": "月租使用者", "Normal": "一般使用者"}

	// 出場show謝謝光臨需除管理卡外更新db出場紀錄
	if viper.GetString("Device.PassMode") == "1" {
		screen.ShowText("謝謝光臨<br>" + identity[userType] + "出場")

	}
	// 非出場show歡迎光臨且一次收費除管理卡外需更新db出場紀錄
	if viper.GetString("Device.PassMode") != "1" {
		screen.ShowText("歡迎光臨<br>" + identity[userType] + "進場")

	}
	// 非自訂evencode才拋監控
	if !strings.Contains(eventCode, "x") {
		monitor.LogCollection(ID, eventCode)
	}

	if _, err := xps.Syslog(viper.GetString("deviceno"), ID, accessTime, eventCode, eventMsg); err == nil {
		checkCarPass()
	}
}

// 不合法程序
func Block(id, eventCode, msg string) {
	screen.ShowText(msg)
	if _, err := xps.Syslog(viper.GetString("deviceno"), id, nowTime(), eventCode, msg); err == nil {
		checkCarPass()
		time.Sleep(time.Second * 3)
	}
}

// 檢查車輛是否通過
func checkCarPass() {
	tStamp := time.Now()
	for {
		if (GetGPIOStat("ACoil") == 0 && GetGPIOStat("BCoil") == 0) || GetGPIOStat("BCoil") == 1 {
			break
		}
		if time.Since(tStamp) >= time.Second*12 {
			tStamp = time.Now()
			screen.ShowText("請盡快通過")
		}
		time.Sleep(time.Second)
	}
}

// 檢查token進場
func CheckToken() bool {
	//卡片Mode||車辨進場失敗
	if GetGPIOStat("Token") == 1 {
		screen.ShowText("歡迎光臨請進場(代幣進場)")
		checkCarPass()
		return true
	}
	return false
}

// local time
func nowTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
