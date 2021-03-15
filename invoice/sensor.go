package invoice

import (
	"fmt"
	"time"
)

const (
	paperEnough = `"\x12"`
	// paperNotEnough = `"\x1e"`
	// paperIsExist   = `"\x92"`
	// paperNotExist  = `"\x12"`
)

func CheckPaperIsExist(pos string) (bool, error) {
	buf := make([]byte, 128)
	for i := 0; i < 2; i++ {
		fd[pos].Write([]byte{
			0x10,
			0x04,
			0x01,
		})
		n, err := fd[pos].Read(buf)
		if err != nil {
			return false, err
		}
		result := fmt.Sprintf("%q", buf[:n])
		if result == paperEnough {
			return true, nil
		}
		time.Sleep(time.Second)
	}
	return false, nil
}

// 檢查紙捲是否足夠
func CheckPaperEnough(pos string) (bool, error) {
	buf := make([]byte, 128)
	for i := 0; i < 2; i++ {
		fd[pos].Write([]byte{
			0x10,
			0x04,
			0x04,
		})
		n, err := fd[pos].Read(buf)
		if err != nil {
			return false, err
		}
		result := fmt.Sprintf("%q", buf[:n])

		if result == paperEnough {
			return true, nil
		}
		time.Sleep(time.Second)
	}
	return false, nil
}
