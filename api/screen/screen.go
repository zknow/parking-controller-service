package screen

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/gogf/gf/os/glog"

	"github.com/spf13/viper"
)

var (
	transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	https = &http.Client{
		Transport: transport,
		Timeout:   time.Second * 3,
	}
)
var nowPage string

func Index()                                     { post("index") }
func LicenseRec()                                { post("license-recognition") }
func LicenseRecFailed()                          { post("license-recognition-fail") }
func SystemProcessing()                          { post("system-processing") }
func Welcome(inTime, balance string)             { post("card-entrance", inTime, balance) }
func FullPark()                                  { post("parking-full") }
func InductionCards()                            { post("induction") }
func InductionCardsFailed()                      { post("induction-fail") }
func Exit(inTime, outTime, cost, balance string) { post("card-export", inTime, outTime, cost, balance) }
func GetCarNumber()                              { post("keyin-license-numbers") }
func AutoLoad()                                  { post("insufficient-balance") }
func InvoiceSaving()                             { post("invoice-saving") }
func InvoiceSaveSuccess()                        { post("invoice-save") }
func UserNotFound()                              { post("induction-no-result") }
func GetCarPlat(pic string)                      { post("license-select", pic) }
func LicenseNotFound()                           { post("license-no-result") }
func UserNotFoundError()                         { post("license-no-result-error") }
func InvoicePrint()                              { post("invoice-print") }
func InvoiceTake()                               { post("invoice-take") }
func ShowText(msg string) {
	post("show-text", msg)
}
func InvoiceFlow() bool {
	// 詢問統編時間
	timeAskTax := viper.GetString("invoice.TaxNoSecond")
	// 歸戶時間
	timeAskReturnProperty := viper.GetString("invoice.PropertySecond")
	// 輸入統邊時間
	timeInputEin := viper.GetString("invoice.TaxNoInput")
	// 是否驗證統邊
	userValidateEin := viper.GetString("invoice.VerifyTaxNo")
	// 是否詢問統邊
	askTaxNo := viper.GetString("invoice.AskTaxNo")
	// 是否歸戶
	askReturnProperty := viper.GetString("invoice.Property")
	return post("invoice-flow", timeAskTax, timeAskReturnProperty, timeInputEin, userValidateEin, askTaxNo, askReturnProperty)
}

func post(param ...string) bool {
	if param[0] == nowPage {
		return true
	}
	defer func() {
		nowPage = param[0]
	}()

	urlPath := "URL.Example"
	action := param[0]
	data := url.Values{"action": {action}}
	switch action {
	case "index":
		name := viper.GetString("invoice.StationName")
		data.Add("pname", name)
	case "card-entrance":
		data.Add("inTime", param[1])
		data.Add("cardBalance", param[2])
	case "card-export":
		data.Add("inTime", param[1])
		data.Add("outTime", param[2])
		data.Add("cost", param[3])
		data.Add("cardBalance", param[4])
	case "invoice-flow":
		data.Add("timeAskEin", param[1])
		data.Add("timeAskReturnProperty", param[2])
		data.Add("timeInputEin", param[3])
		data.Add("userValidateEin", param[4])
		data.Add("askEin", param[5])
		data.Add("askReturnProperty", param[6])
	case "license-select":
		pic := param[1]
		data.Add("sCno", pic)
	case "show-text":
		data.Add("text", param[1])
	default:
		break
	}

	resp, err := http.PostForm(urlPath, data)
	if err != nil {
		return false
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if !strings.Contains(`{"retCode": "1"}`, string(b)) {
		return false
	}
	return true
}
