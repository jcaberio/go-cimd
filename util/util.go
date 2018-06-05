package util

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/paulbellamy/ratecounter"
)

var seqNum int32 = 0
var DeliveryCount uint64 = 0
var SubmitCount uint64 = 0
var TpsCount int64 = 0
var MoMsgChan = make(chan string, 10)
var RateCounter = ratecounter.NewRateCounter(1 * time.Second)

func NextSeqNum() []byte {
	current := atomic.LoadInt32(&seqNum)
	plusTwo := atomic.AddInt32(&seqNum, 2) & 0x00FF
	atomic.StoreInt32(&seqNum, plusTwo)
	n := fmt.Sprintf("%03d", current)
	return []byte(n)
}

// Message is latest message to be displayed in the web UI.
type Message struct {
	Message   string `json:"message"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
}
