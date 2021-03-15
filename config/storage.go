package config

import (
	"errors"

	"github.com/zknow/parkingCharge/api/xps"

	"github.com/spf13/viper"
)

var (
	Deviceno = ""
	ManaList = ""
	Device   = map[string]string{"PassMode": "", "MinAmount": "", "CarType": ""}
	Stage    = map[string]string{"StageName": "", "StagePhone": "", "EmergencyNo": "",
		"BCoilEnable": "", "HeartCardEnable": "", "IsStopMonthly": ""}
	Camera  = map[string]string{"Enable": "", "IP": "", "Account": "", "Password": "", "Channel": ""}
	Proxy   = map[string]string{"IP": "ipaddr.example", "Port": "443", "Account": "api@test.com.tw", "Password": "!@#api===", "BlackList": ""}
	Invoice = map[string]string{
		"EDCId":              "", //EDCID
		"askForInvoice":      "", //缺紙是否申請發票
		"askTaxNo":           "", //是否詢問統邊
		"cutPaper":           "", //是否切紙
		"cutType":            "", //全切半切
		"deviceId":           "", //設備名稱
		"itemName":           "", //品名
		"noPickUpNotice":     "", //未取紙警告
		"paperRemainWarning": "", //印表機切換警告
		"property":           "", //是否詢問歸護
		"propertySecond":     "", //詢問歸護時間
		"sellerTaxNo":        "", //賣方統邊
		"stationName":        "", //廠名
		"taxNoInput":         "", //輸入統邊時間
		"taxNoSecond":        "", //詢問統邊時間
		"verifyTaxNo":        "", //是否驗證統邊
	}
)
var (
	ErrDeviceFail   = errors.New("取得Device參數失敗")
	ErrStageFail    = errors.New("取得Stage參數失敗")
	ErrInvoiceFail  = errors.New("取得Invoice參數失敗")
	ErrCameraFail   = errors.New("取得Camera參數失敗")
	ErrProxyFail    = errors.New("取得Proxy參數失敗")
	ErrManaListFail = errors.New("取得ManaList參數失敗")
)

func GetAll() error {
	if !GetDevice() {
		return ErrDeviceFail
	}
	if !GetStatge() {
		return ErrStageFail
	}
	if !GetCamera() {
		return ErrCameraFail
	}
	if !GetInvoice() {
		return ErrInvoiceFail
	}
	if !GetProxy() {
		return ErrProxyFail
	}
	if !GetManagerList() {
		return ErrManaListFail
	}
	return nil
}
func IsExist(op ...string) bool {
	if len(op) >= 1 {
		if !viper.IsSet(op[0]) {
			return false
		}
		for val := range viper.Get(op[0]).(map[string]string) {
			if val == "" {
				return false
			}
		}
		return true
	}
	return true
}
func GetManagerList() bool {
	m, err := xps.GetParameter(viper.GetString("deviceno"), "manaList")
	if err != nil || m == nil {
		return false
	}
	ManaList = m["manaList"]
	viper.Set("ManaList", ManaList)
	return true
}
func GetDevice() bool {
	m, err := xps.GetParameter(viper.GetString("deviceno"), "device")
	if err != nil || m == nil {
		return false
	}
	viper.Set("Device.PassMode", m["deviceType"])
	viper.Set("Device.CarType", m["carType"])
	viper.Set("Device.MinAmount", m["minAmount"])
	return true
}
func GetStatge() bool {
	m, err := xps.GetParameter(viper.GetString("deviceno"), "stage")
	if err != nil || m == nil {
		return false
	}
	Stage["StageName"] = m["stationName"]
	Stage["StagePhone"] = m["stationPhone"]
	Stage["EmergencyNo"] = m["emergencyContact"]
	Stage["BCoilEnable"] = m["BCoilEnable"]
	Stage["HeartCardEnable"] = m["heartCardEnable"]
	Stage["IsStopMonthly"] = m["isStopMonthly"]
	return true
}
func GetInvoice() bool {
	m, err := xps.GetParameter(viper.GetString("deviceno"), "invoice")
	if err != nil || m == nil {
		return false
	}
	for k, v := range m {
		key := "Invoice." + k
		viper.Set(key, v)
	}
	return true
}
func GetCamera() bool {
	m, err := xps.GetParameter(viper.GetString("deviceno"), "camera")
	if err != nil || m == nil {
		return false
	}
	Camera = map[string]string{"Enable": m["licensePlate"], "IP": m["cameraIp"], "Account": m["cameraAccount"], "Password": m["cameraPassword"], "Channel": m["cameraChannel"]}
	for k, v := range Camera {
		key := "Camera." + k
		viper.Set(key, v)
	}
	return true
}

func GetProxy() bool {
	m, err := xps.GetParameter(viper.GetString("deviceno"), "proxy")
	if err != nil || m == nil {
		return false
	}
	Proxy = map[string]string{"IP": m["proxyIp"], "Port": m["proxyPort"], "Account": m["proxyUserId"], "Password": m["proxyPassword"], "BlackList": m["sftpBlackListName"]}
	for k, v := range Camera {
		key := "Proxy." + k
		viper.Set(key, v)
	}
	return true
}
