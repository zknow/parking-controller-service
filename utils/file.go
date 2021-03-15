package utils

import (
	"io"
	"io/ioutil"
	"os"
	"time"

	log "github.com/gogf/gf/os/glog"
	flock "github.com/theckman/go-flock"
)

func CreateDir(DirPath string) {
	if _, StatErr := os.Stat(DirPath); os.IsNotExist(StatErr) {
		err := os.MkdirAll(DirPath, 0777)
		if err != nil {
			log.Fatal("Create Dir Fail :" + DirPath)
		}
	}
}

func CreateFile(path string) {
	if _, StatErr := os.Stat(path); os.IsNotExist(StatErr) {
		var file, err = os.Create(path)
		if err != nil {
			log.Fatal("Create File Fail :", file)
		}
		file.Close()
	}
}

func DeleteFile(file string) {
	var err = os.Remove(file)
	if err != nil {
		log.Error(err)
	}
	log.Info("已刪除檔案 =>", file)
}

func CopyFile(src, dst string) error {
	in, err := os.OpenFile(src, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Truncate(0)
		in.Close()
	}()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func CheckFileIsExist(path string) bool {
	if _, StatErr := os.Stat(path); os.IsNotExist(StatErr) {
		return false
	}
	return true
}

func GetDirSizeByByte(dir string) int {
	var size int64
	files, _ := ioutil.ReadDir(dir)
	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			size += file.Size()
		}
	}
	return int(size)
}

func GetFileSizeByByte(path string) int {
	info, err := os.Stat(path)
	if err != nil {
		log.Error("Get File Size Failed ", err)
	}
	return int(info.Size())
}

func LockFile(path string, loop bool) *flock.Flock {
	locker := flock.New(path) //lock handler
	for {
		isLocked, err := locker.TryLock()
		if err != nil {
			log.Error("file locking error : ", err)
		}
		if isLocked {
			break
		}
		if !loop {
			return nil
		}
		time.Sleep(time.Millisecond * 500)
	}
	return locker
}
