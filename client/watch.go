package client

import (
	"github.com/daidd2019/ddfetch/model"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"log"
	"net/rpc/jsonrpc"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var g_rpcaddress = "127.0.0.1:8888"
var g_watchdir = "."
var g_filter = regexp.MustCompile("log$")

type FileChanges struct {
	AppName     string
	FileName    string
	Writed      chan bool
	CloseFile   chan bool
	QuitRuntime chan bool
	FileHander  *os.File
	Pos         int64
}

func (fc *FileChanges) CloseFileHander() {
	if fc.FileHander != nil {
		log.Println("close file hander ")
		fc.FileHander.Close()
		fc.FileHander = nil
	}
}

func (fc *FileChanges) ClearChan() {
	log.Println("clear file ", fc.FileName)
	close(fc.Writed)
	close(fc.CloseFile)
	close(fc.QuitRuntime)
}

func (fc *FileChanges) NotifyWrited() {
	sendOnlyIfEmpty(fc.Writed)
}

func (fc *FileChanges) NotifyCloseFile() {
	sendOnlyIfEmpty(fc.CloseFile)
}

func (fc *FileChanges) NotifyQuit() {
	sendOnlyIfEmpty(fc.QuitRuntime)
}

func sendOnlyIfEmpty(ch chan bool) {
	select {
	case ch <- true:
	default:
	}
}

func NewFileChanges(appName, fileName string) *FileChanges {
	fc := FileChanges{appName, fileName, make(chan bool, 1), make(chan bool, 1), make(chan bool), nil, 0}
	go func() {
		conn, err := jsonrpc.Dial("tcp", g_rpcaddress)
		if err != nil {
			conn = nil
		} else {
			defer conn.Close()
		}
		var reply model.Reply
		dirName, fileName := filepath.Split(fc.FileName)
		dirName = strings.Replace(dirName, g_watchdir, "", 1)
		localIp, _ := ExternalIP()

		for {
			select {
			case <-fc.Writed:
				if fc.FileHander == nil {
					log.Println("track file ", fc.FileName)
					var err error
					fc.FileHander, err = os.Open(fc.FileName)
					if fc.Pos != 0 {
						fc.FileHander.Seek(fc.Pos, os.SEEK_SET)
						continue
					}

					if err != nil {
						fc.FileHander = nil
						continue
					}

				}
				if conn != nil {
					buffer := make([]byte, 1024)

					stat, _ := fc.FileHander.Stat()
					size := stat.Size()
					if fc.Pos == 0 && size > 1024 {
						fc.Pos = size
						fc.FileHander.Seek(0, os.SEEK_END)
						continue
					} else if size <= fc.Pos {
						fc.FileHander.Seek(0, os.SEEK_SET)
					}
					readlen, _ := fc.FileHander.Read(buffer)

					currentPos, _ := fc.FileHander.Seek(0, os.SEEK_CUR)
					msg := model.RpcMsg{
						fc.AppName, localIp, dirName, fileName, buffer[:readlen]}
					err := conn.Call("MsgSave.Save", msg, &reply)
					if err != nil {
						log.Println("err", err)
						fc.FileHander.Seek(fc.Pos, os.SEEK_SET)
						conn = nil
					} else {
						fc.Pos = currentPos
					}
				} else {
					log.Println("connect ...", g_rpcaddress)
					conn, _ = jsonrpc.Dial("tcp", g_rpcaddress)
				}

			case <-fc.CloseFile:
				fc.CloseFileHander()
			case <-fc.QuitRuntime:
				fc.CloseFileHander()
				fc.ClearChan()
				return
			}
		}
	}()
	return &fc
}

type AppWatch struct {
	Appname string
	Watcher *fsnotify.Watcher
	Wdone   chan bool
	files   sync.Map
}

func (fm *AppWatch) Init(ip, appname, basedir, filter string) {
	fm.Appname = appname
	g_watchdir = filepath.FromSlash(basedir)
	var err error
	fm.Watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("create watcher error", err)
	}
	fm.Wdone = make(chan bool)
	go func() {
		for {
			select {
			case event := <-fm.Watcher.Events:
				fm.process_event(event)
			case err := <-fm.Watcher.Errors:
				log.Println("watch error:", err)
			}
		}
	}()

	fm.files = sync.Map{}
	g_rpcaddress = ip
	g_filter = regexp.MustCompile(filter)
}

func (fm *AppWatch) process_event(event fsnotify.Event) {

	switch event.Op {
	case fsnotify.Create:
		finfo, err := os.Stat(event.Name)
		if err != nil {
			log.Println("stat file error:", err)
		}
		if finfo.IsDir() {
			fm.Watcher.Add(event.Name)
			log.Println("create watcher:", event.Name)
		}
	case fsnotify.Rename, fsnotify.Remove:
		fm.Watcher.Remove(event.Name)
		log.Println("remove watcher:", event.Name)
		if v, ok := fm.files.Load(event.Name); ok {
			fc := v.(*FileChanges)
			fc.NotifyQuit()
			fm.files.Delete(event.Name)
		}

	case fsnotify.Write:
		if g_filter.MatchString(event.Name) {
			if v, ok := fm.files.Load(event.Name); ok {
				fc := v.(*FileChanges)
				fc.NotifyWrited()
			} else {
				fc := NewFileChanges(fm.Appname, event.Name)
				fc.NotifyWrited()
				fm.files.Store(event.Name, fc)
			}
		}
	}
}

func (fm *AppWatch) walkdir(path string) {
	fm.AddWatcher(path)
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal("opendir:", err)
	}
	for _, fi := range dir {
		fpath := filepath.FromSlash(path + "/" + fi.Name())
		if fi.IsDir() {
			if strings.HasPrefix(fi.Name(), ".") {
				continue
			}
			if strings.HasPrefix(fi.Name(), "..") {
				continue
			}
			if strings.Contains(fi.Name(), "lost+found") {
				continue
			}
			fm.AddWatcher(fpath)
			go fm.walkdir(fpath)
		}
	}
}

func (fm *AppWatch) AddWatcher(path string) {
	err := fm.Watcher.Add(path)
	if err != nil {
		log.Fatal("add watcher error:", err)
	}
	log.Println("add path", path)
}

func (fm *AppWatch) Start() {
	go fm.walkdir(g_watchdir)
	defer fm.Watcher.Close()
	<-fm.Wdone
}
