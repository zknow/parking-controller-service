package adapterClient

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

const addr = "/tmp/adapter.sock"

var (
	ErrChanIsClosed = errors.New("Channel Is Colsed")
)

var (
	conn      net.Conn
	notifyMsg chan string
	screenMsg chan string
)

func init() {
	notifyMsg = make(chan string)
	screenMsg = make(chan string)
}

func CloseAllChan() {
	close(notifyMsg)
	close(screenMsg)
}

//ConnUnixSocketChannel 建立與parkigChargeAdapter Unix Socket的連線和msg通道
func ConnUnixSocketChannel() error {
	var err error
	conn, err = net.Dial("unix", addr)
	if err != nil {
		fmt.Println("連接 Unix Server失敗:", err.Error())
		return err
	}

	go func() {
		for {
			buf := make([]byte, 2048)
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Println("讀取unix server data err:", err.Error())
				conn.Close()
				return
			}
			rsp := string(buf[0:n])
			fmt.Println(rsp)
			if strings.Contains(rsp, "Notify") {
				notifyMsg <- rsp
				continue
			} else if strings.Contains(rsp, "Screen") {
				screenMsg <- rsp
				continue
			}
			time.Sleep(time.Second)
		}
	}()
	return nil
}
