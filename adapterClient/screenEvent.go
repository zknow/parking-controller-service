package adapterClient

import (
	"encoding/json"
	"log"
)

type Screen struct {
	EventName string
	CarNumber string

	CarPlat carPlat
	Invoice invoice
}
type invoice struct {
	TxNumber    string
	SaveInvoice string
}
type carPlat struct {
	Number     string
	AccessTime string
}

func ListenScreenMsg() (*Screen, error) {
	select {
	case e, ok := <-screenMsg:
		if !ok {
			return nil, ErrChanIsClosed
		}
		log.Println(e)
		s := &Screen{}
		json.Unmarshal([]byte(e), s)
		log.Println(s)
		return s, nil
	default:
	}
	return nil, nil
}
