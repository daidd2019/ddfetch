package main

import (
	"flag"
	"github.com/daidd2019/ddfetch/client"
	"runtime"
)

var (
	ddserv   = flag.String("server", "10.247.32.250:8888", "dd serve ip:port")
	app      = flag.String("app", "test", "app name")
	watchdir = flag.String("watch", ".", "track dir")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	filem := new(client.AppWatch)
	filem.Init(*ddserv, *app, *watchdir)
	filem.Start()
}
