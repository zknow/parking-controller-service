package services

import (
	"errors"
	"strconv"
	"strings"

	"github.com/zknow/parkingCharge/api/xps"

	"github.com/spf13/viper"
)

const (
	successCode = "1"
)

// 進出場定義
const (
	passModeIn         = "0"
	passModeOut        = "1"
	passModeOnceCharge = "2"
)

// access code定義
const (
	accessCodeIn  = "1"
	accessCodeOut = "2"
)

// 入場收費類型
const (
	accessTypeCard   = "0" // 卡片
	accessTypeCamara = "1" // 車牌辨識
)

// Verify 合法性驗證結構
type Verify struct {
	Legal     bool
	EventCode string
	Msg       string
	InTime    string
	Amount    string
	Sn        string
}

// 取得用戶類型
func VerifyUserType(id string) (string, error) {
	// 管理員
	if strings.Contains(viper.GetString("managerNoList"), id) {
		return "Manager", nil
	}

	resp, err := xps.CardType(viper.GetString("deviceno"), id, viper.GetString("Device.CarType"))
	if err != nil {
		return "", err
	}
	form := make(map[string]string)
	for k, val := range resp {
		form[k] = val.(string)
	}

	if form["retCode"] != successCode {
		return "", errors.New(form["retMsg"])
	}
	cardVal := form["retVal"]

	// 月租用戶
	if cardVal == "1" {
		return "Monthly", nil
	}
	// 一般用戶
	return "Normal", nil
}

// 驗證是否滿足最低進場金額
func IsPassMinAmountEnought(inBalance string) bool {
	balance, _ := strconv.Atoi(inBalance)
	//餘額小於最低入場金額
	if viper.GetString("Device.PassMode") == passModeIn && balance < viper.GetInt("Device.MinAmount") {
		return false
	}
	return true
}

// 驗證進出場合法性
func VerifyPassLegal(ID, accessTime, Heart, TransferCost, MRTTime string) (*Verify, error) {
	var accessCode string
	var accessType string

	switch viper.GetString("Device.PassMode") {
	case passModeIn:
		accessCode = accessCodeIn
	case passModeOut:
		accessCode = accessCodeOut
	case passModeOnceCharge:
		accessCode = accessCodeOut

	}
	//卡片|車辨模式
	if viper.GetString("Camera.Enable") == accessTypeCamara {
		accessType = accessTypeCamara
	} else {
		accessType = accessTypeCard
	}

	rspMap, err := xps.IsLegal(viper.GetString("deviceno"), ID, accessCode, accessTime, viper.GetString("Device.CarType"), accessType, Heart, TransferCost, MRTTime)
	if err != nil {
		return nil, err
	}

	if rspMap["retCode"].(string) != successCode {
		return &Verify{Legal: false, EventCode: rspMap["eventCode"].(string), Msg: rspMap["retMsg"].(string)}, nil
	}

	result := &Verify{}
	switch viper.GetString("Device.PassMode") {
	case passModeIn:
		result = &Verify{Legal: true, EventCode: rspMap["retVal"].(string), Msg: rspMap["retMsg"].(string)}
	case passModeOut:
		switch val := rspMap["retVal"].(type) {
		case map[string]interface{}:
			verifyedInfo := map[string]string{}
			for key, v := range val {
				verifyedInfo[key] = v.(string)
			}
			result = &Verify{Legal: true, EventCode: verifyedInfo["eventCode"],
				Amount: verifyedInfo["amount"], InTime: verifyedInfo["inTime"],
				Msg: rspMap["retMsg"].(string), Sn: verifyedInfo["sn"]}
		case string:
			result = &Verify{Legal: true, EventCode: "A009",
				Amount: "0", InTime: "2017-10-10 15:20:25",
				Msg: "一般卡出場-成功", Sn: "1"}
		}
	case passModeOnceCharge:
		accessCode = accessCodeOut
		rspMap, err := xps.IsLegal(viper.GetString("deviceno"), ID, accessCode, accessTime, viper.GetString("Device.CarType"), accessType, Heart, TransferCost, MRTTime)
		if err != nil {
			return nil, err
		}

		if rspMap["retCode"].(string) != successCode {
			return &Verify{Legal: false, EventCode: rspMap["eventCode"].(string), Msg: rspMap["retMsg"].(string)}, nil
		}

		switch verifyedInfo := rspMap["retVal"].(type) {
		case map[string]string:
			result = &Verify{Legal: true, EventCode: verifyedInfo["eventCode"],
				Amount: verifyedInfo["amount"], InTime: verifyedInfo["inTime"],
				Msg: rspMap["retMsg"].(string), Sn: verifyedInfo["sn"]}
		case string:
			result = &Verify{Legal: true, EventCode: "A009",
				Amount: "0", InTime: "2017-10-10 15:20:25",
				Msg: "一般卡出場-成功", Sn: "1"}
		}
	default:
		break
	}
	return result, nil
}
