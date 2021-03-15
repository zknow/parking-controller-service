package xps

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	log "github.com/gogf/gf/os/glog"

	"github.com/spf13/viper"
	"github.com/zknow/parkingCharge/utils"
)

var client = &http.Client{Timeout: time.Second * 5}

// 詢問合法性
func IsLegal(deviceno, customerID, accessCode, accessTime, carType, accessType, HeartCard, TransferCost, MRTTime string) (map[string]interface{}, error) {
	data := url.Values{
		"deviceNo":   {deviceno},
		"customerId": {customerID},
		"accessCode": {accessCode},
		"accessTime": {accessTime},
		"carType":    {carType},
		"accessType": {accessType},
	}
	// 出場Post data
	if accessCode == "2" {
		data.Add("heartCard", HeartCard)
		data.Add("transfer", TransferCost)
		data.Add("mrtTime", MRTTime)
	}

	return Post("isLegal", data.Encode())
}

// DbUpdate 更新資料庫
func DbUpdate(deviceno, customerID, accessTime, payType, carType, payID, sn string) (map[string]interface{}, error) {
	data := url.Values{
		"deviceNo":   {deviceno},
		"customerId": {customerID},
		"accessTime": {accessTime},
		"payType":    {payType},
		"carType":    {carType},
		"payCardNo":  {payID},
		"sn":         {sn},
	}
	resp, err := Post("dbUpdate", data.Encode())
	if err != nil || resp["retCode"] != "1" {
		log.Error("dbUpdate api Post failed", err, "start save loss")
		go func() {
			lock := utils.LockFile("./tmp/dbUpdate.lock", true)
			loss, _ := url.QueryUnescape(data.Encode())
			SaveLoss("./tmp/dbUpdate.loss", loss)
			_ = lock.Unlock()
		}()
	}
	return resp, err
}

// 系統事件紀錄
func Syslog(deviceno, customerID, systime, eventCode, msg string) (map[string]interface{}, error) {
	data := url.Values{
		"deviceNo":   {deviceno},
		"customerId": {customerID},
		"systemTime": {systime},
		"eventCode":  {eventCode},
		"message":    {msg},
	}

	resp, err := Post("syslog", data.Encode())
	if err != nil || resp["retCode"] != "1" {
		log.Warning("syslog Post failed", err, "start saveloss")
		go func() {
			lock := utils.LockFile("./tmp/syslog.lock", true)
			loss, _ := url.QueryUnescape(data.Encode())
			SaveLoss("./tmp/syslog.loss", loss)
			_ = lock.Unlock()
		}()
	}
	return resp, err
}

// 判斷卡片種類 月租卡或一般卡
func CardType(deviceno, ID, carType string) (map[string]interface{}, error) {
	data := url.Values{
		"deviceNo":   {deviceno},
		"customerId": {ID},
		"carType":    {carType},
		"systemTime": {nowTime()},
	}
	return Post("cardType", data.Encode())
}

// 取得車牌位置
func GetLicensePicture(deviceno, carNumber string) (string, error) {
	data := url.Values{
		"deviceNo":   {deviceno},
		"customerId": {carNumber},
	}

	rsp, err := Post("getPicture", data.Encode())
	if err != nil {
		return "", err
	}
	jsonString, _ := json.Marshal(rsp["retVal"])
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}

// 取得發票參數
func GetInvoice(deviceno string) (map[string]interface{}, error) {
	data := url.Values{
		"deviceNo": {deviceno},
	}
	resp, err := Post("getInvoice", data.Encode())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return resp, nil
}

// 標記發票上傳
func SaveInvoice(deviceno, invoiceNo, amount, buyerTX, pMark, txTime string) {
	data := url.Values{
		"deviceNo":  {deviceno},
		"invoiceNo": {invoiceNo},
	}
	Post("saveInvoice", data.Encode())
}

// 模糊比對
func LicensePlateLegality(deviceno, carPlat string) (map[string]interface{}, error) {
	data := url.Values{
		"deviceNo":     {deviceno},
		"licensePlate": {carPlat},
	}
	return Post("licensePlateLegality", data.Encode())
}

// 儲存車牌圖片
func SaveLicensePlate(deviceno, licensePlate, accessTime, photoData, mode string) {
	data := url.Values{
		"deviceNo":     {deviceno},
		"licensePlate": {licensePlate},
		"accessTime":   {accessTime},
		"photo":        {"VGhpcyBpcyBYUFMgQVBJIHNhdmVMaWNlbnNlUGxhdGU="},
		"type":         {mode},
	}
	Post("saveLicensePlate", data.Encode())
}

// xps使用的post function
func Post(apiName string, postData string) (map[string]interface{}, error) {
	ip := viper.GetString("stage.serverIp")
	urlPath := "http://" + ip + "/api/" + apiName
	req, err := http.NewRequest("POST", urlPath, strings.NewReader(postData))
	checkErr(err)

	req.Header.Add("captcha", "This is Captcha") //xps vertify
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	val, _ := url.ParseQuery(postData)
	jsonData, err := json.Marshal(val)
	checkErr(err)
	log.Info(apiName, "api post:", string(jsonData))

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	jsonResp := make(map[string]interface{})
	err = json.Unmarshal(body, &jsonResp)
	if err != nil {
		return nil, err
	}
	JSONEncode, err := json.Marshal(jsonResp)
	if err != nil {
		return nil, err
	}

	log.Info(apiName, "api resp:", string(JSONEncode))
	return jsonResp, nil
}
func nowTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func hmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func secret() string {
	const s = "Rand Code"
	decoded, err := hex.DecodeString(s)
	if err != nil {
		log.Error(err)
	}
	return string(decoded)
}

func pictureBase64(name string) (string, error) {
	filename := "pictures/pls/" + name
	FileData, err := ioutil.ReadFile(filename)
	EncodeData := base64.StdEncoding.EncodeToString([]byte(FileData))
	return EncodeData, err
}

// 保存失敗的api data
func SaveLoss(lossPath string, data string) {
	log.Info("Start save data to:", lossPath, "...")
	f, _ := os.OpenFile(lossPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	defer f.Close()
	if _, err := f.WriteString(data + "\n"); err == nil {
		log.Info("Data 成功寫入", lossPath, "unlock done")
	}
}

func checkErr(err error) {
	if err != nil {
		log.Error(err)
	}
}

// 取得後台參數
func GetParameter(deviceno, getType string) (map[string]string, error) {
	data := url.Values{
		"deviceNo": {deviceno},
		"type":     {getType},
	}

	resp, err := Post("getParameter", data.Encode())
	if err != nil {
		return nil, err
	}
	if resp["retCode"].(string) != "1" {
		return nil, nil
	}
	if getType == "manaList" {
		return map[string]string{"manaList": resp["retVal"].(string)}, nil
	}
	val := resp["retVal"].(map[string]interface{})
	m := make(map[string]string)
	for i, v := range val {
		m[i] = v.(string)
	}
	return m, nil
}

// 更新票證參數至後台
func TicketOperation(deviceno, cardType string, js interface{}) bool {
	bs, err := json.Marshal(js)
	if err != nil {
		return false
	}
	data := url.Values{
		"deviceNo": {deviceno},
		"type":     {cardType},
		"field":    {string(bs)},
	}
	resp, err := Post("ticketOperation", data.Encode())
	if err != nil {
		return false
	}
	if resp["retCode"].(string) != "1" {
		return false
	}
	return true
}
