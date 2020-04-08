package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/rpc"
	"os"
	"strconv"
	"time"

	"github.com/yqjoe/projsync/projsync/confmgr"
	"github.com/yqjoe/projsync/projsync/proto"
	"github.com/yqjoe/projsync/projsynctool/server"
)

func main() {
	if err := confmgr.Init(); err != nil {
		fmt.Println("confmgr.Init fail")
		return
	}

	client, err := rpc.Dial("tcp", ":6547")
	if err != nil {
		fmt.Println("Dial fail")
		return
	}

	fmt.Println("tool start:", os.Args)

	if len(os.Args) < 4 {
		fmt.Printf("Usage:%v projectname optype taskname", os.Args[0])
		return
	}

	projectname := os.Args[1]
	optype := os.Args[2]

	if optype == "conf" {
		confname := os.Args[3]
		if confname == "getautoclose" {
			req := &proto.ReqGetAutoClose{}
			req.ProjectName = projectname
			var rsp proto.RspGetAutoClose
			err := client.Call("TaskServer.GetAutoClose", req, &rsp)
			if err != nil {
				fmt.Println("err:", err.Error())
			} else {
				fmt.Println("autoclose:", rsp.AutoClose)
			}
		} else if confname == "openautoclose" {
			req := &proto.ReqSetAutoClose{}
			req.ProjectName = projectname
			req.AutoClose = true
			var rsp proto.RspSetAutoClose
			err := client.Call("TaskServer.SetAutoClose", req, &rsp)
			if err != nil {
				fmt.Println("openautoclose fail")
			} else {
				fmt.Println("openautoclose succ")
			}
		} else if confname == "closeautoclose" {
			req := &proto.ReqSetAutoClose{}
			req.ProjectName = projectname
			req.AutoClose = false
			var rsp proto.RspSetAutoClose
			err := client.Call("TaskServer.SetAutoClose", req, &rsp)
			if err != nil {
				fmt.Println("closeautoclose fail")
			} else {
				fmt.Println("closeautoclose succ")
			}
		} else if confname == "getautoinfo" {
			req := &proto.ReqAutoInfo{}
			req.ProjectName = projectname
			var rsp proto.RspAutoInfo
			err := client.Call("TaskServer.GetAutoInfo", req, &rsp)
			if err != nil {
				fmt.Println("getautoinfo fail")
			} else {
				fmt.Println("project[", rsp.ProjectName, "] AutoClose[", rsp.AutoClose, "] AutoLen[", len(rsp.AutoInfoArray), "]")
				for _, autoinfo := range rsp.AutoInfoArray {
					fmt.Println("task[", autoinfo.TaskName, "] Circle[", autoinfo.AutoCircle,
						"Minute] LastDoAutoStamp[", time.Unix(autoinfo.LastDoTimestamp, 0).String(), "]")
				}
			}
		}
	} else if optype == "task" {
		taskname := os.Args[3]
		conf := confmgr.GetTaskConf(projectname, taskname)
		if conf == nil {
			fmt.Println("ProjectName or TaskName not impl")
			return
		}

		// Add Task
		req := &proto.ReqAddTask{}
		err = initReqAddTask(req, projectname, taskname, os.Args)
		if err != nil {
			fmt.Println("err:", err.Error())
		}

		var rsp proto.RspAddTask
		if conf.TaskPrinter == "yes" {
			closechan := make(chan int, 1)

			req.SyncToolPrintSvrAddr = genLocalAttr()
			atcall := client.Go("TaskServer.AddTask", req, &rsp, nil)

			//  wait remote call goroutine
			go func() {
				rspcall := <-atcall.Done
				if rspcall.Error != nil {
					closechan <- 1
					return
				}

				if rsp.Ret != 0 {
					fmt.Println("add task fail, err:", rsp.Err)
					closechan <- 1
				}
			}()

			// printersvr goroutine
			printersvr := server.NewPrinterServer(req.SyncToolPrintSvrAddr, closechan)
			go printersvr.Serve()

			// close
			<-closechan
		} else { // no
			req.SyncToolPrintSvrAddr = ""
			//client.Call("TaskServer.AddTask", req, &rsp)
			client.Go("TaskServer.AddTask", req, &rsp, nil)
		}
	} else {
		fmt.Printf("Wrong optype")
	}
}

func genLocalAttr() string {
	rand.Seed(time.Now().UnixNano())
	return (":" + strconv.Itoa(10000+rand.Intn(9999)))
}

func initReqAddTask(req *proto.ReqAddTask, projectname, taskname string, args []string) error {
	req.ProjectName = projectname
	req.TaskName = taskname
	// TODO 这里需要优化
	switch taskname {
	case "savefile":
		if len(args) < 5 {
			return errors.New("Less Args")
		}
		req.Putstepfile = args[4]
	default:
		if len(args) >= 5 {
			for i := 4; i < len(args); i++ {
				req.CallParams = append(req.CallParams, args[i])
			}
		}
	}
	return nil
}
