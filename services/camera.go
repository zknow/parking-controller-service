package services

import (
	"github.com/zknow/parkingCharge/api/camera"
	"github.com/zknow/parkingCharge/api/xps"

	"github.com/spf13/viper"
)

const (
	compareMode_different = "0"
	compareMode_Equal     = "1"
)

// 取得車號
func GetCarLicenseInfo() (string, string, error) {
	id, err := camera.GetInfo()
	if err != nil {
		return "", "", err
	}
	//模糊比對
	comparedId, msg, err := CompareCarLicence(id)
	if err != nil {
		return "", "", err
	}
	return comparedId, msg, nil
}

// 模糊比對
func CompareCarLicence(carNumber string) (string, string, error) {
	// 模糊比對API
	rsp, err := xps.LicensePlateLegality(viper.GetString("deviceno"), carNumber)
	if err != nil {
		return "", "", err
	}
	if rsp["retCode"].(string) != "1" {
		return "", "設備異常", nil
	}

	retVal := rsp["retVal"].(map[string]interface{})
	licensePlate := retVal["customerId"].(string)
	checkMode := retVal["mode"].(string)
	// pass mode out
	if viper.GetString("Device.PassMode") == passModeOut {
		if checkMode == compareMode_different { //無相同車牌
			msg := "找不到場內車牌<br>" + licensePlate
			return "", msg, nil
		}
		return licensePlate, "", nil
	}
	//pass mode in or once
	if checkMode == compareMode_Equal {
		msg := "重複進場<br>" + licensePlate
		return "", msg, nil
	}
	return licensePlate, "", nil
}
