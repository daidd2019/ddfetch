package main

import (
	"flag"
	"github.com/daidd2019/ddfetch/server"
)

var (
	port      = flag.String("port", "8888", "server port")
	keep      = flag.String("keep", "/tmp", "keep path")
	clearTime = flag.Int64("cleartime", 30, "clear file max time")
)

func main() {
	flag.Parse()
	server.Start(*port, *keep, *clearTime)
}
