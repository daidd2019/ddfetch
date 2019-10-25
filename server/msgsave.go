package server

import (
	"github.com/daidd2019/ddfetch/model"
	"log"
	"os"
	"sync"
	"time"
)

type MsgSave struct {
	KeepDir  string
	MapFiles sync.Map
}

type FileInfo struct {
	key           string
	FileHander    *os.File
	LastWriteTime int64
}

func NewFileInfo(app, basepath, subpath, filename, host string) *FileInfo {

	key := app + host + subpath + filename
	filedir := basepath + "/" + app + "/" + host + subpath
	fullpath := filedir + filename

	if exist, err := PathExists(filedir); err == nil {
		if !exist {
			os.MkdirAll(filedir, os.ModePerm)
		}
	}

	writer, err := os.OpenFile(fullpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("error...", err)
		writer = nil
	}
	return &FileInfo{key, writer, time.Now().Unix()}

}

func NewMsgSave(keepDir string, fileCloseTime int64) *MsgSave {
	if exist, err := PathExists(keepDir); err == nil {
		if !exist {
			os.Mkdir(keepDir, os.ModePerm)
		}
	} else {
		log.Fatal("check", keepDir)
	}

	msg := &MsgSave{keepDir, sync.Map{}}
	go func() {
		for {
			time.Sleep(60 * time.Second)
			now := time.Now().Unix()
			msg.MapFiles.Range(func(k, v interface{}) bool {
				fileInfo := v.(*FileInfo)
				if now-fileInfo.LastWriteTime >= fileCloseTime {
					fileInfo.FileHander.Close()
					log.Println("close file", fileInfo.key)
					msg.MapFiles.Delete(k)
					return false
				}
				return true
			})
		}
	}()
	return msg
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (msgsave *MsgSave) Save(msg model.RpcMsg, reply *model.Reply) error {
	key := msg.AppName + msg.HostIp + msg.DirName + msg.FileName

	vd, ok := msgsave.MapFiles.Load(key)
	if ok {
		//exists
		fileItem := vd.(*FileInfo)
		n, err := fileItem.FileHander.Write(msg.FileContent)
		if err != nil {
			log.Println("write error", err)
		} else {
			log.Println("Write length ", n)
		}

		fileItem.LastWriteTime = time.Now().Unix()
	} else {
		filenew := NewFileInfo(msg.AppName, msgsave.KeepDir, msg.DirName, msg.FileName, msg.HostIp)
		n, err := filenew.FileHander.Write(msg.FileContent)
		if err != nil {
			log.Println("write error", err)
		} else {
			log.Println("Write length ", n)
		}

		msgsave.MapFiles.Store(key, filenew)
		log.Println("create ", key)
	}

	reply.Msg = "ok"
	reply.Code = 0
	return nil
}
