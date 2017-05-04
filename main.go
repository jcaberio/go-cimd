package main

import (
	"fmt"
	"log"
	"net"

	"github.com/jcaberio/go-cimd/cimd"
	"github.com/jcaberio/go-cimd/view"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()
}

func main() {
	port := fmt.Sprint(":", viper.GetInt("port"))
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
		if err == nil {
			pdu.Decode()
		}
	}
}
