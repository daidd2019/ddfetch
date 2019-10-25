package main

import (
	"github.com/daidd2019/ddfetch/model"
	"net/rpc/jsonrpc"
	"log"
)


func CallRpc(msg model.RpcMsg, reply *model.Reply) error {
	conn, err := jsonrpc.Dial("tcp", "10.247.32.250:8888")
	if err != nil {
		log.Fatalln("dialing error", err)
		return err
	}
	defer conn.Close()
	err = conn.Call("MsgSave.Save", msg, &reply)
	return err
}


func main() {

	msg := model.RpcMsg{
		AppName : "test",
		HostIp : "127.0.0.1",
		DirName : "/data/go/src/github.com/daidd2019",
		FileName : "msg.go",
		FileContent : []byte("dddddddddddddddddddddddfffffffffffffffffff"),
	}

	var reply model.Reply

	err := CallRpc(msg, &reply)
	if err != nil {
		log.Println("err", err)

	} else {
		log.Println("reply msg", reply.Msg, "reply code", reply.Code)
	}
}
