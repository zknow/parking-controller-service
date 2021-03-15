package ticketClient

import (
	"net"

	log "github.com/gogf/gf/os/glog"
)

func createTicketConnection() (net.Conn, error) {
	var err error
	sockPath := "/tmp/ticket.sock"
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		log.Println("[Error] Reader UnixSocket:", err)
		return nil, err
	}
	return conn, nil
}
func send(src []byte) (string, error) {
	conn, err := createTicketConnection()
	if conn == nil {
		return "", err
	}
	_, err = conn.Write(src)
	if err != nil {
		log.Fatal("Reader UnixSocket:", err)
	}
	buf := make([]byte, 1024)
	c, err := conn.Read(buf)
	if err != nil {
		log.Error(err)
	}
	return string(buf[0:c]), nil
}
