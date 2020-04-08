package main

import (
	"fmt"

	"github.com/yqjoe/projsync/projsync/confmgr"
	"github.com/yqjoe/projsync/projsync/server"
)

func main() {
	if err := confmgr.Init(); err != nil {
		fmt.Println("confmgr.Init fail")
		return
	}

	fmt.Println("projsync svr running")

	server.RunTaskServer()
}
