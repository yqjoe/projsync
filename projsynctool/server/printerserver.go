package server

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/yqjoe/projsync/projsync/proto"
)

type PrinterServer struct {
	svraddr   string
	rpcserver RpcPrinterServer
}

func NewPrinterServer(svraddr string, c chan int) *PrinterServer {
	svr := &PrinterServer{}
	svr.svraddr = svraddr
	svr.rpcserver.closechan = c
	return svr
}

func (server *PrinterServer) Serve() error {
	printersvr := rpc.NewServer()
	printersvr.Register(&server.rpcserver)
	l, err := net.Listen("tcp", server.svraddr)
	if err != nil {
		fmt.Println("Listen fail")
		return err
	}

	printersvr.Accept(l)
	return nil
}

type RpcPrinterServer struct {
	closechan chan int
}

func (server *RpcPrinterServer) PrintTaskInfo(req *proto.ReqPrintTaskInfo, rsp *proto.RspPrintTaskInfo) error {
	fmt.Printf("%v", req.Info)
	return nil
}

func (server *RpcPrinterServer) ClosePrinterSvr(req *proto.ReqClosePrinterSvr, rsp *proto.RspClosePrinterSvr) error {
	server.closechan <- 1
	return nil
}
