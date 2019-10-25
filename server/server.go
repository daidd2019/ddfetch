package server

import (
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

func Start(port, keepPath string, fileCloseTime int64) {

	if err := rpc.Register(NewMsgSave(keepPath, fileCloseTime)); err != nil {
		panic(nil)
	}
	address := "0.0.0.0:" + port

	if listener, err := net.Listen("tcp", address); err != nil {
		panic(nil)
	} else {
		log.Println("listener address", address)
		for {
			if conn, err := listener.Accept(); err != nil {
				log.Println(err)
				continue
			} else {
				go jsonrpc.ServeConn(conn)
			}
		}
	}
}
