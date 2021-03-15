package utils

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/gogf/gf/os/glog"
	"github.com/spf13/viper"
)

const (
	BYTE = 1.0 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
)

func LogRotate(dir string) {
	filePath := dir + "park.log"

	//如果不存在,不壓縮
	if !CheckFileIsExist(filePath) {
		return
	}
	t := time.Now().Format("20060102")
	err := exec.Command("tar", "-zcf", dir+t+".tar.gz", filePath).Run()
	if err != nil {
		log.Error(err)
	}
	os.Remove(dir + "park.log")
}

func deleteTarGz(dir string) {
	skillfolder := dir
	var rmFileName = "99999999"
	files, _ := ioutil.ReadDir(skillfolder)
	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			filename := file.Name()
			if strings.Contains(filename, ".tar.gz") {
				tmpFileName := strings.Split(filename, ".tar.gz")[0]
				if tmpFileName < rmFileName {
					rmFileName = filename
				}
			}
		}
	}
	if rmFileName == "99999999" {
		return
	}
	err := os.Remove(dir + rmFileName)
	if err != nil {
		log.Fatal(err)
	}
}

func DeleteOldLog(dir string) {
	maxSize := viper.GetInt("stage.logMaxSize")
	// log.Info("size", float64(maxSize)/MEGABYTE, "MB")
	for {
		if GetDirSizeByByte(dir) > maxSize*KILOBYTE {
			deleteTarGz(dir)
		} else {
			break
		}

		time.Sleep(time.Millisecond)
	}
}
