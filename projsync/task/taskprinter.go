package task

import (
	//"fmt"
	"net/rpc"

	"github.com/yqjoe/projsync/projsync/proto"
)

type msgID int

const (
	msgIDContent msgID = 1
	msgIDClose   msgID = 2
)

type msgPack struct {
	msgid   msgID
	content string
}

// TaskPrinter is a printer of task
type TaskPrinter struct {
	client *rpc.Client
	buffer chan msgPack
}

func NewTaskPrinter(printersvraddr string) *TaskPrinter {
	tp := &TaskPrinter{nil, nil}
	if len(printersvraddr) == 0 {
		// 不远程发送
		return tp
	}

	var err error
	tp.client, err = rpc.Dial("tcp", printersvraddr)
	if err != nil {
		tp.client = nil
		return tp
	}

	tp.buffer = make(chan msgPack, 10)

	go func() {
		for {
			msg := <-tp.buffer
			if msg.msgid == msgIDClose {
				var req proto.ReqClosePrinterSvr
				var rsp proto.RspClosePrinterSvr
				tp.client.Call("RpcPrinterServer.ClosePrinterSvr", req, &rsp)
				break
			} else if msg.msgid == msgIDContent {
				req := &proto.ReqPrintTaskInfo{}
				req.Info = msg.content
				var rsp proto.RspPrintTaskInfo
				tp.client.Call("RpcPrinterServer.PrintTaskInfo", req, &rsp)
			}
		}
	}()

	return tp
}

func (printer *TaskPrinter) Write(p []byte) (n int, err error) {
	if printer.client != nil {
		printer.buffer <- msgPack{msgIDContent, string(p)}
	} //else {
	//fmt.Printf("%v", string(p))
	//}
	return len(p), nil
}

func (printer *TaskPrinter) Close() {
	if printer.client != nil {
		printer.buffer <- msgPack{msgIDClose, ""}
	}
}
