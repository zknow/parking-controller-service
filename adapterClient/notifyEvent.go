package adapterClient

import (
	"encoding/json"
	"log"
	"time"
)

type Notify struct {
	EventName  string
	DevStatus  string
	Gate       string
	UpdateList string
	Emergency  string
	Counter888 counter888
}
type counter888 struct {
	Option string `json:"option"`
	Count  string `json:"count"`
}

func ListenNotifyMsg() (*Notify, error) {
	select {
	case e, ok := <-notifyMsg:
		if !ok {
			return nil, ErrChanIsClosed
		}
		log.Println("Get event = ", e)
		x := &Notify{}
		err := json.Unmarshal([]byte(e), x)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(x)
		return x, nil
	default:
	}
	return nil, nil
}

var (
	stop     = make(chan bool)
	openning = false
)

//三種開門模式任務
func openTask(op string) {
	switch op {
	case "open":
		openGate()
	case "keepOpen":
		log.Println("KeepOpen")
		if openning {
			return
		}
		keepOpenGate()
	case "close":
		log.Println("close")
		if !openning {
			return
		}
		stop <- true
	}
}

func keepOpenGate() {
	openning = true
	for {
		openGate()
		time.Sleep(time.Second * 2)
		select {
		case <-stop:
			openning = false
			return
		default:
		}
	}
}
func openGate() {
	log.Println("open")
}
