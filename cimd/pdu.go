package cimd

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type PDU struct {
	Conn   net.Conn
	CmdID  []byte
	SeqNum []byte
	Data   map[string]string
}

func NewPDU(c net.Conn) (*PDU, error) {
	reader := bufio.NewReader(c)
	raw, _ := reader.ReadSlice(ETX)
	if len(raw) == 0 {
		return nil, errors.New("Empty read on connection")
	}
	pduParts := bytes.Split(raw, []byte{TAB})[:]
	pduPartsETXOmitted := pduParts[:len(pduParts)-1]
	data := make(map[string]string)
	for _, pduPart := range pduPartsETXOmitted {
		paramClnVal := string(pduPart[:])
		paramValSlice := strings.Split(paramClnVal, ":")
		param := paramValSlice[0]
		val := paramValSlice[1]
		data[param] = val
	}
	rawHeader := pduPartsETXOmitted[0]
	header := bytes.TrimFunc(rawHeader, func (c rune) bool {return c == STX || c == ETX})
	cmdAndSeq := bytes.Split(header, []byte{COLON})
	cmdID := cmdAndSeq[0]
	seqNum := cmdAndSeq[1]

	return &PDU{
		CmdID:  cmdID,
		SeqNum: seqNum,
		Data:   data,
		Conn:   c,
	}, nil
}

func (p *PDU) Decode() {
	switch string(p.CmdID[:]) {
	case LOGIN:
		log.Println("LOGIN COMMAND")
		p.LoginResp(p.Authenticate())

	case ALIVE:
		log.Println("ALIVE COMMAND")
		p.KeepAlive()
	case LOGOUT:
		log.Println("LOGOUT COMMAND")
		p.LogoutResp()

	case SUBMIT_MSG:
		log.Println("SUBMIT_MESSAGE COMMAND")
		p.SubmitMessageResp(p.SubmitMessage())

	default:
		log.Println("UNKNOWN COMMAND")
		p.UnknownCmd()
	}
}

func (p *PDU) LoginResp(b bool) {
	if b {
		byteToWrite := make([]byte, 0)
		byteToWrite = append(byteToWrite, STX)
		byteToWrite = append(byteToWrite, LOGIN_RESP...)
		byteToWrite = append(byteToWrite, COLON)
		byteToWrite = append(byteToWrite, p.SeqNum...)
		byteToWrite = append(byteToWrite, TAB)
		byteToWrite = append(byteToWrite, ETX)
		p.Conn.Write(byteToWrite)
		log.Println("VALID LOGIN")
	} else {
		byteToWrite := make([]byte, 0)
		byteToWrite = append(byteToWrite, STX)
		byteToWrite = append(byteToWrite, GENERAL_ERROR_RESP...)
		byteToWrite = append(byteToWrite, COLON)
		byteToWrite = append(byteToWrite, p.SeqNum...)
		byteToWrite = append(byteToWrite, TAB)
		byteToWrite = append(byteToWrite, ERROR_CODE...)
		byteToWrite = append(byteToWrite, COLON)
		byteToWrite = append(byteToWrite, INVALID_LOGIN...)
		byteToWrite = append(byteToWrite, TAB)
		byteToWrite = append(byteToWrite, ETX)
		p.Conn.Write(byteToWrite)
		log.Println("INVALID LOGIN")
	}
}

func (p *PDU) LogoutResp() {
	byteToWrite := make([]byte, 0)
	byteToWrite = append(byteToWrite, STX)
	byteToWrite = append(byteToWrite, LOGOUT_RESP...)
	byteToWrite = append(byteToWrite, COLON)
	byteToWrite = append(byteToWrite, p.SeqNum...)
	byteToWrite = append(byteToWrite, TAB)
	byteToWrite = append(byteToWrite, ETX)
	p.Conn.Write(byteToWrite)
}

func (p *PDU) Authenticate() bool {
	userIdentity := p.Data[USER_IDENTITY]
	password := p.Data[PASSWORD]
	ui := viper.GetString("cimd_user")
	pw := viper.GetString("cimd_pw")
	log.Println("username: ", ui)
	log.Println("password: ", pw)
	return userIdentity == ui && password == pw
}

func (p *PDU) KeepAlive() {
	byteToWrite := make([]byte, 0)
	byteToWrite = append(byteToWrite, STX)
	byteToWrite = append(byteToWrite, KEEP_ALIVE_RESP...)
	byteToWrite = append(byteToWrite, COLON)
	byteToWrite = append(byteToWrite, p.SeqNum...)
	byteToWrite = append(byteToWrite, TAB)
	byteToWrite = append(byteToWrite, ETX)
	p.Conn.Write(byteToWrite)
}

func (p *PDU) SubmitMessage() bool {
	log.Println("DST_ADDR: ", p.Data[DST_ADDR])
	log.Println("ORIG_ADDR: ", p.Data[ORIG_ADDR])
	log.Println("USER_DATA: ", p.Data[USER_DATA])
	return p.Data[DST_ADDR] != "" && p.Data[ORIG_ADDR] != "" && p.Data[USER_DATA] != ""

}

func (p *PDU) SubmitMessageResp(b bool) {
	if b {
		byteToWrite := make([]byte, 0)
		byteToWrite = append(byteToWrite, STX)
		byteToWrite = append(byteToWrite, SUBMIT_MSG_RESP...)
		byteToWrite = append(byteToWrite, COLON)
		byteToWrite = append(byteToWrite, p.SeqNum...)
		byteToWrite = append(byteToWrite, TAB)
		byteToWrite = append(byteToWrite, DST_ADDR_RESP...)
		byteToWrite = append(byteToWrite, COLON)
		byteToWrite = append(byteToWrite, []byte(p.Data[DST_ADDR])...)
		byteToWrite = append(byteToWrite, TAB)
		byteToWrite = append(byteToWrite, SVC_CENTER_RESP...)
		byteToWrite = append(byteToWrite, COLON)
		byteToWrite = append(byteToWrite, SVC_CENTER_TIMESTAMP...)
		byteToWrite = append(byteToWrite, TAB)
		byteToWrite = append(byteToWrite, ETX)
		p.Conn.Write(byteToWrite)
	}
}

func (p *PDU) UnknownCmd() {
	byteToWrite := make([]byte, 0)
	byteToWrite = append(byteToWrite, STX)
	byteToWrite = append(byteToWrite, GENERAL_ERROR_RESP...)
	byteToWrite = append(byteToWrite, COLON)
	byteToWrite = append(byteToWrite, p.SeqNum...)
	byteToWrite = append(byteToWrite, TAB)
	byteToWrite = append(byteToWrite, ERROR_CODE...)
	byteToWrite = append(byteToWrite, COLON)
	byteToWrite = append(byteToWrite, UNEXPECTED_OPERATION...)
	byteToWrite = append(byteToWrite, TAB)
	byteToWrite = append(byteToWrite, ETX)
	p.Conn.Write(byteToWrite)
}