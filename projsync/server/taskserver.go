package server

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/yqjoe/projsync/projsync/confmgr"
	"github.com/yqjoe/projsync/projsync/proto"
	"github.com/yqjoe/projsync/projsync/task"
)

type taskKey struct {
	ProjectName, TaskName string
}

type taskMap map[taskKey]*task.Task

type TaskServer struct {
	currTask taskMap

	taskauto *TaskAuto
}

func (server *TaskServer) AddTask(req *proto.ReqAddTask, rsp *proto.RspAddTask) error {
	if server.isTaskExist(req.ProjectName, req.TaskName) {
		rsp.Ret = -1
		rsp.Err = "task already exist"
		return nil
	}

	onetask := task.NewTask(req.ProjectName, req.TaskName)
	server.addTaskMap(req.ProjectName, req.TaskName, onetask)

	// TODO 这里需要优化
	switch req.TaskName {
	case "savefile":
		onetask.SetPutStepFile(req.Putstepfile)
	default:
		if len(req.CallParams) > 0 {
			onetask.SetCallParams(req.CallParams)
		}
	}
	if len(req.SyncToolPrintSvrAddr) > 0 {
		onetask.SetSyncToolPrintSvrAddr(req.SyncToolPrintSvrAddr)
	}
	onetask.InitTaskFromConf()
	onetask.Run()

	server.delTaskMap(req.ProjectName, req.TaskName)
	return nil
}

func (server *TaskServer) GetAutoClose(req *proto.ReqGetAutoClose, rsp *proto.RspGetAutoClose) error {
	rsp.ProjectName = req.ProjectName
	rsp.AutoClose = server.taskauto.GetAutoClose(req.ProjectName)
	return nil
}

func (server *TaskServer) SetAutoClose(req *proto.ReqSetAutoClose, rsp *proto.RspSetAutoClose) error {
	server.taskauto.SetAutoClose(req.ProjectName, req.AutoClose)
	return nil
}

func (server *TaskServer) GetAutoInfo(req *proto.ReqAutoInfo, rsp *proto.RspAutoInfo) error {
	rsp.ProjectName = req.ProjectName
	rsp.AutoInfoArray, rsp.AutoClose = server.taskauto.GetAutoInfo(req.ProjectName)
	return nil
}

func (server *TaskServer) isTaskExist(projectname, taskname string) bool {
	key := taskKey{projectname, taskname}
	if _, ok := server.currTask[key]; ok {
		return true
	}

	return false
}

func (server *TaskServer) addTaskMap(projectname, taskname string, task *task.Task) {
	server.currTask[taskKey{projectname, taskname}] = task
}

func (server *TaskServer) delTaskMap(projectname, taskname string) {
	delete(server.currTask, taskKey{projectname, taskname})
}

func RunTaskServer() {
	tasksvr := new(TaskServer)
	tasksvr.currTask = make(taskMap)
	svr := rpc.NewServer()
	svr.Register(tasksvr)

	l, err := net.Listen("tcp", confmgr.GetTaskServerAddr())
	if err != nil {
		fmt.Println("Listen fail")
		return
	}

	taskauto := NewTaskAuto(tasksvr)
	taskauto.GoServe()
	tasksvr.taskauto = taskauto

	svr.Accept(l)
}
