package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/zknow/parkingCharge/adapterClient"
	"github.com/zknow/parkingCharge/services"

	"github.com/zknow/parkingCharge/api/screen"
	"github.com/zknow/parkingCharge/config"
	"github.com/zknow/parkingCharge/controllers"
	"github.com/zknow/parkingCharge/invoice"
	"github.com/zknow/parkingCharge/utils"

	log "github.com/gogf/gf/os/glog"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"
	flock "github.com/theckman/go-flock"
)

var (
	mainLock  *flock.Flock
	logFile   *os.File
	devStatus string
)

// 服務模式介面
type serviceInterface interface {
	Serve()
	Reset()
}

// gpio信號
const (
	gpio_on  = 1
	gpio_off = 0
)

const (
	devStatusPath = "/tmp/devStatus"
	mainLockPath  = "/tmp/main.lock"
)

func Init() {
	if mainLock = utils.LockFile(mainLockPath, false); mainLock == nil {
		log.Fatal("程式正在執行!!")
	}

	initNeedDirectory()
	initLogger()

	if !utils.CheckFileIsExist(devStatusPath) {
		err := ioutil.WriteFile(devStatusPath, []byte("alive"), os.ModePerm)
		if err != nil {
			log.Fatal("初始化裝置狀態檔案失敗")
		}
		devStatus = "alive"
	} else {
		bs, err := ioutil.ReadFile(devStatusPath)
		if err != nil {
			devStatus = "stop"
		} else {
			devStatus = string(bs)
		}
	}

	config.LoadConfig()

	// 讀卡機初始化 + 取得deviceID
	if ok := services.InitialReader(); !ok {
		log.Fatal("請先啟用/確認讀卡socket存在")
	}
	if devno, ok := services.GetDeviceID(); !ok {
		log.Warning("取得deviceno失敗，設定替代deviceno : " + viper.GetString("deviceno"))
		viper.Set("deviceno", viper.GetString("deviceno"))
	} else {
		viper.Set("deviceno", devno)
	}

	// 確認設備初始狀態
	if !utils.CheckFileIsExist(devStatusPath) {
		err := ioutil.WriteFile(devStatusPath, []byte("alive"), os.ModePerm)
		if err != nil {
			log.Fatal("初始化裝置狀態檔案失敗")
		}
	}

	// 初始化發票機 & 設定紙捲狀態
	if invoice.InitPrinter() {
		invoice.SetPaperStatus()
	}

	// Log 打包排程
	go utils.MakeSchedule(func() {
		utils.LogRotate("logs/")
		initLogger()
	})

	// Log 壓縮檔案排程
	go utils.Ticker(func() {
		utils.DeleteOldLog("logs/")
	}, time.Minute, true)

	go utils.Ticker(func() {
		lossMemberDataUpdate()
	}, time.Minute, true)
}

func main() {
	// 初始化必要參數,併發,路徑.....
	Init()

	defer closeAll()

	if err := config.GetAll(); err != nil {
		log.Error(err)
	}

	if err := adapterClient.ConnUnixSocketChannel(); err != nil {
		log.Fatal(adapterClient.ErrChanIsClosed)
	}

	// 監聽kill信號,程式被殺掉時能夠安全退出
	safelyQuit()

	// 開始服務流程
	start()
}

// 安全退出信號監聽
func safelyQuit() {
	log.Info("Start listen quit signal ...")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Println(sig)
		closeAll()
		os.Exit(1)
	}()
}

//開始服務程序
func start() {
	park := getController()
	if park == nil {
		log.Fatal("初始化Controller失敗!")
	}
	log.Println("開始服務 Type:", reflect.TypeOf(park))
	for {
		//監聽事件
		event, err := adapterClient.ListenNotifyMsg()
		if err != nil {
			log.Warning("與 unixSocket 斷線")
			adapterClient.ConnUnixSocketChannel()
			time.Sleep(time.Second)
			continue
		}
		if event != nil {
			switch event.EventName {
			case "DevStatus": //設備狀態
				devStatus = event.DevStatus
			case "UpdateList": //更新參數
				list := event.UpdateList
				switch true {
				case strings.Contains("ManaList", list):
					config.GetManagerList()
				case strings.Contains("Device", list):
					config.GetDevice()
				case strings.Contains("Stage", list):
					config.GetStatge()
				case strings.Contains("Camera", list):
					config.GetCamera()
				case strings.Contains("Proxy", list):
					config.GetProxy()
				case strings.Contains("Invoice", list):
					config.GetInvoice()
				}
			case "Counter888":
				// 計數器
			case "Emergency":
				// 請洽管理員 + 電話
			case "Gate":
				// 遠端開門
			}
		}

		if !strings.Contains(devStatus, "alive") {
			screen.ShowText("停止服務...")
			log.Info("devStatus : " + devStatus)
			time.Sleep(time.Second)
			continue
		}

		switch services.GetGPIOStat("ACoil") { //檢查A線圈
		case gpio_on: // 開始服務(依parkType)
			if services.GetGPIOStat("BCoil") == gpio_on { //前方有車loop
				park.Reset()
				log.Println("前方有車請稍後")
				screen.ShowText("前方有車請稍後")
			}
			park.Serve()
		case gpio_off: //回待機並重製所有狀態
			screen.Index()
			park.Reset()
		default:
			log.Fatal("gpio get failed")
		}

		time.Sleep(time.Second)
	}
}

//取得服務模式
func getController() serviceInterface {
	switch services.CheckParkMode() {
	case "cardsIn":
		return &controllers.CardsIn{}
	case "cardsOut":
		return &controllers.CardsOut{}
	case "cardsOnce":
		return &controllers.CardsOnce{}
	case "cameraIn":
		return &controllers.CameraIn{}
	case "cameraOut":
		return &controllers.CameraOut{}
	default:
		return nil
	}
}

func initLogger() {
	var err error
	logFilePath := "logs/park.log"
	logFile.Close()
	logFile, err = os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Init Logger Failed!")
	}

	logWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetWriter(logWriter)
	log.SetFlags(log.F_TIME_STD | log.F_FILE_SHORT)
}

// 備份遺失的Member資料
func lossMemberDataUpdate() {
	services.LossDataUpdate("dbUpdate")
	services.LossDataUpdate("syslog")
}

func initNeedDirectory() {
	path := []string{
		"logs",
		"pictures",
		"tmp",
	}
	for i := 0; i < len(path); i++ {
		utils.CreateDir(path[i])
	}
}

func closeAll() {
	logFile.Close()
	mainLock.Unlock()
}
