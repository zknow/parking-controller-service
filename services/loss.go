package services

import (
	"bufio"
	"io"

	log "github.com/gogf/gf/os/glog"

	"net/url"
	"os"
	"strings"
	"time"

	"github.com/zknow/parkingCharge/api/xps"
	"github.com/zknow/parkingCharge/utils"
)

// 補傳遺失的卡片(用戶)資訊
func LossDataUpdate(lossName string) {
	path := "./tmp/" + lossName
	tmpPath := "./tmp/" + lossName + ".tmp"
	lossLock := "./tmp/" + lossName + ".lock"

	lock := utils.LockFile(lossLock, false)
	if lock == nil {
		return
	}
	defer func() {
		_ = lock.Unlock()
	}()

	if !utils.CheckFileIsExist(path) {
		return
	}
	// 複製 .loss to .tmp
	log.Info("開始複製", path, "資料至", tmpPath, "...")
	if utils.CopyFile(path, tmpPath) != nil {
		log.Warning("copy", path, "error")
		return
	}
	i := 0
	// 清除loss file並上傳  失敗的寫入tmp
	arrMap := parseLossFile(tmpPath)
	for _, query := range arrMap {
		postData, _ := url.ParseQuery(query)
		resp, err := xps.Post(lossName, postData.Encode())
		if resp["retCode"] != "1" || err != nil {
			i++
			xps.SaveLoss(path, query)
		}
	}
	time.Sleep(time.Millisecond * 200)
	os.Remove(tmpPath)
	if i == 0 {
		os.Remove(path)
	}
}

// Parse Loss檔
func parseLossFile(path string) []string {
	f, _ := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	defer f.Close()

	r := bufio.NewReader(f)

	var arr []string
	for {
		data, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else {
			arr = append(arr, strings.Trim(data, "\n"))
		}
	}
	_ = f.Truncate(0)
	return arr
}
