package invoice

import (
	"time"

	log "github.com/gogf/gf/os/glog"

	"github.com/zknow/parkingCharge/api/xps"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/tarm/serial"
)

// 狀態變數
var (
	fd        = map[string]*serial.Port{}
	Status    = map[string]string{} //0:缺紙 , 1:有紙 , 2:無設備
	NowStatus string
)

// 錯誤訊息
const (
	ErrTwoPinterFailed     = "雙印表機錯誤"
	ErrTwoPinterOutOfPaper = "雙印表機紙捲耗盡"
)

// 錯誤訊息
const (
	Normal  = "0"
	Enable  = "1"
	Disable = "2"
)

// 空發票資訊
type Inv struct {
	ShopName      string
	SellerAddress string
	InvoiceNo     string
	SellerTaxNo   string
	RandomNo      string
	QRCode        string
}

// 初始化印表機
func InitPrinter() bool {
	var err error
	c := &serial.Config{Name: "/dev/ttymxc3", Baud: 9600, ReadTimeout: time.Second * 5}
	fd["PrinterA"], err = serial.OpenPort(c)
	if err != nil {
		log.Error("初始化印表機A失敗")
		Status["PrinterA"] = Disable
	}

	c = &serial.Config{Name: "/dev/ttymxc2", Baud: 115200, ReadTimeout: time.Second * 5}
	fd["PrinterB"], err = serial.OpenPort(c)
	if err != nil {
		log.Error("初始化印表機B失敗")
		Status["PrinterB"] = Disable
	}

	if Status["PrinterA"] == Disable && Status["PrinterB"] == Disable {
		NowStatus = ErrTwoPinterFailed
		return false
	}
	return true
}

// 取得紙捲狀態
func SetPaperStatus() {
	if NowStatus == ErrTwoPinterFailed {
		return
	}
	ok, err := CheckPaperEnough("PrinterA")
	if err != nil {
		Status["PrinterA"] = Disable
	}
	if ok {
		Status["PrinterA"] = Enable
		NowStatus = "PrinterA"
	} else {
		Status["PrinterA"] = Normal
	}

	ok, err = CheckPaperEnough("PrinterB")
	if err != nil {
		Status["PrinterB"] = Disable
	}
	if ok {
		Status["PrinterB"] = Enable
		NowStatus = "PrinterB"
	} else {
		Status["PrinterB"] = Normal
	}

	if Status["PrinterA"] == Normal && Status["PrinterB"] == Normal {
		NowStatus = "PrinterA"
	}
}

// 發票流程(統編,歸戶,金額)
func InvProcess(buyerTX, amount string) bool {
	NowStatus = "PrinterB"
	rsp, err := xps.GetInvoice(viper.GetString("deviceno"))
	if err != nil {
		return false
	}
	if rsp["retCode"].(string) != "1" {
		return false
	}
	inv := &Inv{}
	err = mapstructure.Decode(rsp["retVal"], inv)
	if err != nil {
		return false
	}
	PrintInv(inv)
	return true
}
