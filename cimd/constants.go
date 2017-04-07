package cimd

import (
	"fmt"
	"sync/atomic"
	"time"
)

const (
	STX                      = 2
	ETX                      = 3
	TAB                      = 9
	COLON                    = 58
	LOGIN                    = "01"
	USER_IDENTITY            = "010"
	PASSWORD                 = "011"
	ALIVE                    = "40"
	LOGOUT                   = "02"
	SUBMIT_MSG               = "03"
	DST_ADDR                 = "021"
	ORIG_ADDR                = "023"
	USER_DATA                = "033"
	SUCCESSFUL_DELIVERY      = "4"
	DELIVER_STAT_REPORT_RESP = "73"
)

var (
	seqNum                  int32 = 0
	DELIVER_STAT_REPORT_REQ       = []byte{50, 51}
	LOGIN_RESP                    = []byte{53, 49}
	LOGOUT_RESP                   = []byte{53, 50}
	GENERAL_ERROR_RESP            = []byte{57, 56}
	ERROR_CODE                    = []byte{57, 48, 48}
	INVALID_LOGIN                 = []byte{49, 48, 48}
	UNEXPECTED_OPERATION          = []byte{49}
	KEEP_ALIVE_RESP               = []byte{57, 48}
	SUBMIT_MSG_RESP               = []byte{53, 51}
	DST_ADDR_RESP                 = []byte{48, 50, 49}
	SVC_CENTER_RESP               = []byte{48, 54, 48}
	SVC_CENTER_TIMESTAMP          = []byte(time.Now().Format("20060102150405"))
	STATUS_CODE                   = []byte{48, 54, 49}
	DISCHARGE_TIME                = []byte{48, 54, 51}
)

func NextSeqNum() []byte {
	current := atomic.LoadInt32(&seqNum)
	plusTwo := atomic.AddInt32(&seqNum, 2) & 0x00FF
	atomic.StoreInt32(&seqNum, plusTwo)
	n := fmt.Sprintf("%03d", current)
	return []byte(n)
}
