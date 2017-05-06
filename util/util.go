package util

import (
	"fmt"
	"sync/atomic"
)

var seqNum int32 = 0
var DeliveryCount uint64 = 0
var SubmitCount uint64 = 0
var MoMsgChan = make(chan string)

func NextSeqNum() []byte {
	current := atomic.LoadInt32(&seqNum)
	plusTwo := atomic.AddInt32(&seqNum, 2) & 0x00FF
	atomic.StoreInt32(&seqNum, plusTwo)
	n := fmt.Sprintf("%03d", current)
	return []byte(n)
}
