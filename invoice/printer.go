package invoice

import (
	"fmt"
	"strings"
)

func PrintInv(inv *Inv) {
	/* 將發票資訊寫入Printer */
	cut("full")
}
func barCode() {
	/* 將barCode資訊寫入Printer */
}
func qrCode() {
	/* 將QrCode資訊寫入Printer */
}

func cut(cmd string) {
	/* 剪裁發票指令 */
}

func writeByte(data []byte) {
	fd[NowStatus].Write(data)
}

func writeString(v ...interface{}) {
	var strconv []string
	for _, item := range v {
		s := fmt.Sprintf("%v", item)
		strconv = append(strconv, s)
	}
	justString := strings.Join(strconv, "")

	b, _ := Encodebig5([]byte(justString))
	fd[NowStatus].Write(b)
}
