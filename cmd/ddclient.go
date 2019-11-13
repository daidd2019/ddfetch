package main

import (
	"flag"
	"github.com/daidd2019/ddfetch/client"
	"runtime"
)

var (
	ddserv   = flag.String("server", "127.0.0.1:8888", "dd serve ip:port")
	app      = flag.String("app", "test", "app name")
	watchdir = flag.String("watch", ".", "track dir")
	filter   = flag.String("filter", ".*", "file filter regexp like log$|txt$ default all file")
	buffsize = flag.Int("buffsize", 4096, "buffer size for send data")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	filem := new(client.AppWatch)
	filem.Init(*ddserv, *app, *watchdir, *filter, *buffsize)
	filem.Start()
}
