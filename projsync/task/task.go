package task

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yqjoe/projsync/projsync/cmd"
	"github.com/yqjoe/projsync/projsync/confmgr"
	//"runtime"
)

type Task struct {
	ProjectName string
	TaskName    string

	cmdlist     []cmd.ICmd
	putstepfile string // winscp put 命令 同步的文件，本地文件的全路径

	callParams []string // winscp call 命令的参数，如果参数不为空，依次填到这里

	// sync tool 打印服务器地址，如果没有则为空
	synctoolprintsvraddr string
	taskprinter          *TaskPrinter
}

func NewTask(projname, taskname string) *Task {
	return &Task{
		projname,
		taskname,
		make([]cmd.ICmd, 0),
		"",
		make([]string, 0),
		"",
		nil}
}

func (task *Task) Run() {
	// 连接synctool print svr
	task.taskprinter = NewTaskPrinter(task.synctoolprintsvraddr)
	// 关闭printer
	defer task.taskprinter.Close()

	for _, icmd := range task.cmdlist {
		cmd.ExecCmd(icmd, task.taskprinter)
	}

	//fmt.Println("goroutine num:", runtime.NumGoroutine())
}

func (task *Task) SetPutStepFile(file string) {
	task.putstepfile = file
}

func (task *Task) SetCallParams(params []string) {
	task.callParams = append(task.callParams, params...)
}

func (task *Task) SetSyncToolPrintSvrAddr(addr string) {
	task.synctoolprintsvraddr = addr
}

func (task *Task) InitTaskFromConf() {
	taskconfobj := confmgr.GetTaskConf(task.ProjectName, task.TaskName)
	if taskconfobj == nil {
		fmt.Println("ProjectTaskConf Not found", taskconfobj.TaskName)
		return
	}

	for _, cmd := range taskconfobj.Cmd {
		task.addTaskCmd(&cmd)
	}
}

func (task *Task) addTaskCmd(cmdconf *confmgr.CmdConf) {
	switch cmdconf.CmdName {
	case "winscp":
		task.addTaskWinScpCmd(cmdconf)
	case "svn":
		task.addTaskSvnCmd(cmdconf)
	case "xcopy":
		task.addTaskXcopyCmd(cmdconf)
	default:
		fmt.Println("cmd not impl:", cmdconf.CmdName)
	}
}

func (task *Task) addTaskWinScpCmd(cmdconf *confmgr.CmdConf) {
	projconf := confmgr.GetProjectConf(task.ProjectName)
	if nil == projconf {
		return
	}

	scpcmd := cmd.NewWinScpCmd()
	scpcmd.SetUser(projconf.User)
	scpcmd.SetPassword(projconf.Password)
	scpcmd.SetHost(projconf.Host)
	scpcmd.SetPort(projconf.Port)

	for _, stepconf := range cmdconf.Step {
		task.addTaskWinScpStep(scpcmd, &stepconf)
	}

	task.cmdlist = append(task.cmdlist, scpcmd)
}

func (task *Task) addTaskWinScpStep(scpcmd *cmd.WinScpCmd, stepconf *confmgr.StepConf) {
	switch stepconf.StepName {
	case "put":
		task.addTaskWinScpPutStep(scpcmd, stepconf)
	case "sync":
		task.addTaskWinScpSyncStep(scpcmd, stepconf)
	case "call":
		task.addTaskWinScpCallStep(scpcmd, stepconf)
	default:
		fmt.Println("Step Not Impl:", stepconf.StepName)
	}
}

func (task *Task) addTaskWinScpPutStep(scpcmd *cmd.WinScpCmd, stepconf *confmgr.StepConf) {
	projconf := confmgr.GetProjectConf(task.ProjectName)
	if nil == projconf {
		return
	}

	putstep := cmd.NewWinScpStepPutFile()
	putstep.SetLocalfile(task.putstepfile)
	putstep.SetRemotefile(task.genRemotefileFromLocalfile(task.putstepfile))
	scpcmd.AddWinScpStep(putstep)
}

func FormatRemotePath(path string) string {
	// char "\" to "/"
	return strings.Replace(path, "\\", "/", -1)
}

func FormatLocalPath(path string) string {
	// char "/" to "\"
	return strings.Replace(path, "/", "\\", -1)
}

func (task *Task) genRemotefileFromLocalfile(localfile string) string {
	projconf := confmgr.GetProjectConf(task.ProjectName)
	if nil == projconf {
		return ""
	}

	localdirlen := len(projconf.Localdir)
	remotefile := projconf.Remotedir + localfile[localdirlen:]
	return FormatRemotePath(remotefile)
}

func (task *Task) addTaskWinScpSyncStep(scpcmd *cmd.WinScpCmd, stepconf *confmgr.StepConf) {
	projconf := confmgr.GetProjectConf(task.ProjectName)
	if nil == projconf {
		return
	}

	syncstep := cmd.NewWinScpStepSync()
	if len(stepconf.OverrideLocaldir) > 0 {
		syncstep.SetLocalDir(FormatLocalPath(stepconf.OverrideLocaldir))
	} else {
		syncstep.SetLocalDir(FormatLocalPath(projconf.Localdir + stepconf.Relativedir))
	}
	if len(stepconf.OverrideRemotedir) > 0 {
		syncstep.SetRemoteDir(FormatRemotePath(stepconf.OverrideRemotedir))
	} else {
		syncstep.SetRemoteDir(FormatRemotePath(projconf.Remotedir + stepconf.Relativedir))
	}
	if stepconf.SyncDirection == "local2remote" {
		syncstep.SetDirection(cmd.WIN_SCP_SYNC_DIRECTION_LOCAL_TO_REMOTE)
	} else {
		syncstep.SetDirection(cmd.WIN_SCP_SYNC_DIRECTION_REMOTE_TO_LOCAL)
	}
	for _, include := range stepconf.Include {
		if len(include) > 0 {
			syncstep.AddInclude(include)
		}
	}
	for _, exclude := range stepconf.Exclude {
		if len(exclude) > 0 {
			syncstep.AddExclude(exclude)
		}
	}

	scpcmd.AddWinScpStep(syncstep)
}

func (task *Task) addTaskWinScpCallStep(scpcmd *cmd.WinScpCmd, stepconf *confmgr.StepConf) {
	projconf := confmgr.GetProjectConf(task.ProjectName)
	if nil == projconf {
		return
	}

	callstep := cmd.NewWinScpStepCall()
	for _, shellcmd := range stepconf.ShellCmd {
		for i := 0; i < len(task.callParams); i++ {
			place_holder := "${Param" + strconv.Itoa(i+1) + "}"
			shellcmd = strings.ReplaceAll(shellcmd, place_holder, task.callParams[i])
		}

		// windows path
		shellcmd = strings.ReplaceAll(shellcmd, "\\", "\\\\")

		//fmt.Println("shellcmd:", shellcmd)
		callstep.AddShellCmd(shellcmd)
	}

	scpcmd.AddWinScpStep(callstep)
}

// svn task
func (task *Task) addTaskSvnCmd(cmdconf *confmgr.CmdConf) {
	projconf := confmgr.GetProjectConf(task.ProjectName)
	if nil == projconf {
		return
	}

	for _, stepconf := range cmdconf.Step {
		svncmd := cmd.NewSvnCmd()
		svncmd.SetOp(stepconf.StepName)
		svncmd.SetSvnDir(projconf.Localdir + "\\" + stepconf.Relativedir)
		svncmd.SetUser(projconf.SvnUser)
		svncmd.SetPassword(projconf.SvnPassword)

		task.cmdlist = append(task.cmdlist, svncmd)
	}
}

// xcopy task
func (task *Task) addTaskXcopyCmd(cmdconf *confmgr.CmdConf) {
	projconf := confmgr.GetProjectConf(task.ProjectName)
	if nil == projconf {
		return
	}

	for _, stepconf := range cmdconf.Step {
		xcopycmd := cmd.NewXcopyCmd()
		xcopycmd.SetSrcFiles(projconf.Localdir + "\\" + stepconf.Relativedir + "\\" + stepconf.FileName)
		xcopycmd.SetDstDir(projconf.Localdir + "\\" + stepconf.DstRelativedir)

		task.cmdlist = append(task.cmdlist, xcopycmd)
	}
}
