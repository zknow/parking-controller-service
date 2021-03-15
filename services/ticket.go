package services

import (
	"os"
	"strconv"
	"time"

	log "github.com/gogf/gf/os/glog"

	"github.com/spf13/viper"
	"github.com/zknow/parkingCharge/api/screen"
	"github.com/zknow/parkingCharge/ticketClient"
)

// 初始化讀卡機
func InitialReader() bool {
	comport := os.Getenv("ReadPort")
	log.Info("COMPORT = ", comport)
	ok := ticketClient.SetPort("ecc", comport)
	if !ok {
		log.Error("Set Reader Port Error")
	}
	proxy := make(map[string]string)
	ok = ticketClient.SetProxy("ecc", "01", proxy)
	if !ok {
		log.Error("Set SetProxy Error")
	}
	ok = ticketClient.GetAndSetBillList("ecc")
	if !ok {
		log.Error("Set List Error")
	}
	return true
}

// 取得設備id
func GetDeviceID() (string, bool) {
	id, ok := ticketClient.GetDeviceID("ecc")
	if !ok {
		log.Error("Get GetDeviceID Error")
		return "", false
	}
	return id, true
}

// 檢查餘額
func CheckCardsBalanceEnought(balance, autoload, amount string) bool {
	cardAmount, _ := strconv.Atoi(balance)
	verifyAmount, _ := strconv.Atoi(amount)
	if cardAmount < verifyAmount && autoload != "Y" {
		return false
	}
	return true
}

// 卡片扣款流程
func CardsDeduct(amount string) (bool, string) {
	return ticketClient.Deduct("1")
}

// ID,Balance,Heart,Autoload
func GetCardInfo() (map[string]string, error) {
	var ok = 1
	cards := []string{}
	if viper.GetInt("cardsEnable.ecc") == ok {
		cards = append(cards, "ecc")
	} else if viper.GetInt("cardsEnable.ipass") == ok {
		cards = append(cards, "ipass")
	} else if viper.GetInt("cardsEnable.icash") == ok {
		cards = append(cards, "icash")
	} else if viper.GetInt("cardsEnable.happycash") == ok {
		cards = append(cards, "happycash")
	}
	var err error
	var rtn map[string]string
	for _, cardtype := range cards {
		rtn = ticketClient.Read(cardtype)
		if rtn != nil {
			rtn["PayType"] = cardtype
			break
		}
	}
	return rtn, err
}

// 刷卡提示
func InductionCardsHint(fullCardStat int) {
	if viper.GetString("Device.PassMode") != "1" && fullCardStat == 1 {
		screen.FullPark()
		return
	}
	screen.InductionCards()
	time.Sleep(time.Second)
}
