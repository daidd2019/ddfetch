package model

import (
	"encoding/json"
	"log"
)

type RpcMsg struct {
	AppName     string `json:"appname"`
	HostIp      string `json:"hostip"`
	DirName     string `json:"dirname"`
	FileName    string `json:"filename"`
	FileContent []byte `json:"filecontent"`
}

type Reply struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
}

func DisplayMsg(msg *RpcMsg) {
	content, err := json.Marshal(msg)
	if err != nil {
		log.Println("json coding err:", err)
	} else {
		log.Println("RpcMsg:", string(content))
	}
}
