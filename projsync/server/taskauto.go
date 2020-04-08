package server

import (
	"time"

	"github.com/yqjoe/projsync/projsync/confmgr"
	"github.com/yqjoe/projsync/projsync/proto"
)

type taskTime struct {
	projectname, taskname string
	lastdotimestamp       int64
	autodocircle          int
}

type taskTimeList []taskTime

type TaskAuto struct {
	timelist taskTimeList

	// tasksvr
	tasksvr *TaskServer

	// projectname to autoclose map
	// 是否关闭自动任务
	autoclosemap map[string]bool
}

func NewTaskAuto(tasksvr *TaskServer) *TaskAuto {
	return &TaskAuto{make(taskTimeList, 0), tasksvr, make(map[string]bool)}
}

func (auto *TaskAuto) GoServe() {
	// 初始化
	auto.init()

	go func() {
		for {
			nowtime := time.Now().Unix()
			for index, tasktime := range auto.timelist {
				autoclose, ok := auto.autoclosemap[tasktime.projectname]
				if ok && autoclose == true {
					continue
				}

				if nowtime <= tasktime.lastdotimestamp+int64(tasktime.autodocircle*60) {
					continue
				}

				tasktime.lastdotimestamp = nowtime
				auto.timelist[index] = tasktime

				// addtask
				auto.addTask(tasktime.projectname, tasktime.taskname)
			}
			time.Sleep(5 * time.Second)
		}
	}()
}

func (auto *TaskAuto) SetAutoClose(projectname string, autoclose bool) {
	auto.autoclosemap[projectname] = autoclose
}

func (auto *TaskAuto) GetAutoClose(projectname string) bool {
	autoclose, ok := auto.autoclosemap[projectname]
	if ok {
		return autoclose
	}

	return false
}

func (auto *TaskAuto) GetAutoInfo(projectname string) ([]proto.AutoInfo, bool) {
	autoinfoarray := make([]proto.AutoInfo, 0)
	for _, tasktime := range auto.timelist {
		if tasktime.projectname != projectname {
			continue
		}

		autoinfo := proto.AutoInfo{tasktime.taskname, tasktime.autodocircle, tasktime.lastdotimestamp}
		autoinfoarray = append(autoinfoarray, autoinfo)
	}

	return autoinfoarray, auto.autoclosemap[projectname]
}

func (auto *TaskAuto) addTask(projectname, taskname string) {
	req := proto.ReqAddTask{}
	req.ProjectName = projectname
	req.TaskName = taskname
	rsp := proto.RspAddTask{}
	go auto.tasksvr.AddTask(&req, &rsp)
}

func (auto *TaskAuto) init() {
	for projectname, projconf := range confmgr.ConfObj.ProjectMap {
		for _, taskconf := range projconf.Task {
			if taskconf.AutoDoTaskCircle > 0 {
				tasktime := taskTime{projectname, taskconf.TaskName, time.Now().Unix(), taskconf.AutoDoTaskCircle}
				auto.timelist = append(auto.timelist, tasktime)
			}

			if taskconf.AutoDoTask == true {
				auto.autoclosemap[projectname] = false
			} else {
				auto.autoclosemap[projectname] = true
			}
		}
	}
}
