package camera

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/gogf/gf/os/glog"

	"github.com/spf13/viper"
)

var client = &http.Client{
	Timeout: time.Second * 3,
}

var (
	ErrNetConnFailed     = errors.New("連線異常")
	ErrScanLicenseFailed = errors.New("車牌辨識失敗")
	ErrGetInfoFailed     = errors.New("get license info failed IP:" + camIP)
	ErrDLPicFailed       = errors.New("DownLoad license picture failed IP:" + picIP)

	info  = make(map[string]string)
	camIP = ""
	picIP = ""

	PhotoName string
	UTC       string
)

//GetInfo 取得車牌資訊
func GetInfo() (string, error) {
	contentType := "application/x-www-form-urlencoded"
	data := "Username=" + viper.GetString("Camera.Password") + "&Password=" + viper.GetString("Camera.Password") + "&UTC=0&TimeOut=2&LatestOne=1"
	log.Info("CarPlatInfoAPI Post :", data)
	resp, err := client.Post(camIP, contentType, strings.NewReader(data))
	if err != nil {
		log.Error("CarPlatInfoAPI api post failed", err)
		return "", ErrNetConnFailed
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("CarPlatInfoAPI", err)
		return "", ErrScanLicenseFailed
	}
	defer resp.Body.Close()

	log.Info("CarPlatInfoAPI Respones :", string(body))

	var m map[string]interface{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		log.Error("CarPlatInfoAPI api respond json Unmarshal failed", err)
		return "", ErrScanLicenseFailed
	}

	dataMap := m["data"].([]interface{}) //object of array using->[]interface{}
	if len(dataMap) < 1 {
		return "", ErrScanLicenseFailed
	}

	indexMap := dataMap[0].(map[string]interface{}) //only one object index = 0
	for i, v := range indexMap {
		if s, ok := v.(string); ok {
			info[i] = s
		}
	}
	return info["RecognizedPlateID"], nil
}

// 下載車牌圖片
func PictureDownload(number, accesstime string) error {
	data := "Username=" + viper.GetString("Camera.Account") +
		"&Password=" + viper.GetString("Camera.Password") +
		"&UTC=" + info["UTC"] +
		"&PhotoName=" + info["PhotoName"]
	contentType := "application/x-www-form-urlencoded"
	resp, err := client.Post(picIP, contentType, strings.NewReader(data))
	if err != nil {
		return err
	}

	log.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return ioutil.WriteFile("pictures/"+accesstime+".jpg", body, os.ModePerm)
}
