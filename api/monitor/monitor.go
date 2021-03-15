package monitor

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	log "github.com/gogf/gf/os/glog"
)

var (
	tr     = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client = &http.Client{Transport: tr, Timeout: time.Second * 5}
)

func LogCollection(cardID, evenCode string) {
	systime := time.Now().Format("20060102150405")

	form := url.Values{
		"Device_ID":   {"XXXXXX"},
		"Shop_name":   {"XXXXXX"},
		"System_time": {systime},
		"Card_SN":     {cardID},
		"API_Code":    {evenCode},
		"Note":        {""},
	}

	ip := ""
	rsp, err := client.PostForm(ip, form)
	if err != nil {
		log.Error("監控API異常", err)
	}
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Error("監控rsp讀取異常", err)
	}
	if string(body) != "OK" {
		log.Error("監控 log_collection.php api 失敗")
	}
}
