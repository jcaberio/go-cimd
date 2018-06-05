package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/jcaberio/go-cimd/cimd"
	"github.com/jcaberio/go-cimd/view"
	"github.com/jcaberio/go-cimd/util"
	"github.com/spf13/viper"
	"github.com/pkg/profile"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()
}

func main() {
	go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
	defer profile.Start(profile.MemProfile).Stop()
	port := fmt.Sprint("0.0.0.0:", viper.GetInt("port"))
	welcomeMsg := fmt.Sprint(viper.GetString("greeting"))
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Println(err)
	}
	go view.Render()
	for {
		conn, err := ln.Accept()
		defer conn.Close()
		if err != nil {
			log.Println(err)
		}
		go handleConnection(conn)
		conn.Write([]byte(welcomeMsg))
	}
}

func handleConnection(conn net.Conn) {
	for {
		pdu, err := cimd.NewPDU(conn)
		if err != nil {
			return
		}
		select {
			case moMessage := <-util.MoMsgChan:
				pdu.DeliverMessage1(moMessage)
			default:
				pdu.Decode()	
		}		
	}
}
