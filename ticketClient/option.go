package ticketClient

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/gogf/gf/os/glog"
)

var nowReaderPos = -1

var Proxy = &struct {
	ProxyIp           string
	ProxyPort         string
	ProxyUserId       string
	ProxyPassword     string
	SftpBlackListName string
}{}

var CardType string

func SetPort(cardType string, port string) bool {
	Switch1509(cardType)
	builder := map[string]interface{}{
		"Cmd":  "SetPort",
		"Port": port,
		"Type": cardType,
	}
	cmd, _ := json.Marshal(builder)
	rsp, err := send(cmd)
	if err != nil {
		return false
	}
	if !strings.Contains(rsp, `"retCode":"1"`) {
		return false
	}
	return true
}

func GetDeviceID(cardType string) (string, bool) {
	deviceM := map[string]string{}
	Switch1509(cardType)
	m := map[string]string{}
	builder := map[string]interface{}{"Cmd": "DeviceID", "Type": cardType}
	cmd, _ := json.Marshal(builder)

	rsp, err := send(cmd)
	if err != nil {
		log.Error("設備異常 Reader Socket Error:", err)
		return "", false
	}
	if !strings.Contains(rsp, `"retCode":"1"`) {
		return "", false
	}
	_ = json.Unmarshal([]byte(rsp), &m)
	deviceM[cardType] = m["retVal"]

	if v, ok := deviceM["ecc"]; ok {
		return v, true
	}
	if v, ok := deviceM["ipass"]; ok {
		return v, true
	}
	if v, ok := deviceM["happyCash"]; ok {
		return v, true
	}
	if v, ok := deviceM["icash"]; ok {
		return v, true
	}
	return "", false
}

// 通知票證取得總帳參數
func GetAndSetBillList(cardType string) bool {
	Switch1509(cardType)
	builder := map[string]interface{}{"Cmd": "List", "Type": cardType}
	cmd, _ := json.Marshal(builder)

	rsp, err := send(cmd)
	if err != nil {
		log.Fatal("設備異常 Reader Socket Error:", err)
	}
	if !strings.Contains(rsp, `"retCode":"1"`) {
		log.Error(cardType, "List Bill failed:", rsp)
		return false
	}
	return true
}

func SetProxy(cardType, tmid string, proxy map[string]string) bool {
	Switch1509(cardType)

	builder := map[string]interface{}{
		"Cmd":   "SetProxy",
		"Type":  cardType,
		"Proxy": proxy,
		"TmID":  tmid,
	}
	cmd, _ := json.Marshal(builder)

	rsp, err := send(cmd)
	if err != nil {
		log.Fatal("設備異常 Reader Socket Error:", err)
	}
	if !strings.Contains(rsp, `"retCode":"1"`) {
		log.Fatal("reader init ", cardType, "failed:", rsp)
	}
	return true
}

func GetParam(cardType string) map[string]string {
	builder := map[string]string{
		"Cmd":  "GetParam",
		"Type": cardType,
	}
	cmd, _ := json.Marshal(builder)
	rsp, err := send(cmd)
	if err != nil {
		log.Fatal("設備異常 Reader Socket Error:", err)
	}
	ok := `"retCode":"1"`
	if !strings.Contains(rsp, ok) {
		return nil
	}
	arr := []interface{}{}
	_ = json.Unmarshal([]byte(rsp), &arr)
	v := arr[0].(map[string]interface{})
	retVal, _ := json.Marshal(v["retVal"])

	m := map[string]string{}
	_ = json.Unmarshal(retVal, &m)
	return m
}

func Read(cardType string) map[string]string {
	builder := map[string]string{"Cmd": "Read", "Type": cardType}
	Switch1509(cardType)
	cmd, _ := json.Marshal(builder)
	rsp, err := send(cmd)
	if err != nil {
		log.Fatal("設備異常 Reader Socket Error:", err)
	}
	if strings.Contains(rsp, `"retCode":"1"`) {
		buf := &struct {
			RetCode string            `json:"retCode"`
			RetVal  map[string]string `json:"retVal"`
		}{}
		_ = json.Unmarshal([]byte(rsp), buf)
		return buf.RetVal
	}
	return nil
}

// 扣款
func Deduct(amount string) (bool, string) {
	builder := map[string]string{
		"Cmd":    "Deduct",
		"Amount": amount,
		"Type":   CardType,
	}
	cmd, _ := json.Marshal(builder)
	rsp, err := send(cmd)
	if err != nil {
		log.Fatal("設備異常 Reader Socket Error:", err)
	}

	ok := `"retCode":"1"`
	if !strings.Contains(rsp, ok) {
		return false, ""
	}
	m := make(map[string]string)
	err = json.Unmarshal([]byte(rsp), &m)
	if err != nil {
		log.Error("扣款解碼失敗", err)
		return false, ""
	}
	return true, m["Balance"]
}

// 切換讀卡機
func Switch1509(card string) {
	if card == "ecc" && nowReaderPos != 0 {
		fmt.Println("1509 switch ecc card")
		nowReaderPos = 0
	}
	if card != "ecc" && nowReaderPos != 1 {
		fmt.Println("1509 switch other card")
		nowReaderPos = 1
	}
}
