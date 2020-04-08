package cmd

///// IWinScpStep
type IWinScpStep interface {
	GetStepArgs() []string
}

///// WinScpCmd
type WinScpCmd struct {
	user     string
	password string
	host     string
	port     string

	// winscp step
	steps []IWinScpStep
}

func NewWinScpCmd() *WinScpCmd {
	return &WinScpCmd{}
}

func (cmd *WinScpCmd) GetCmdName() string {
	return "WinScp"
}

func (cmd *WinScpCmd) GetCmdArgs() []string {
	args := make([]string, 0)
	args = append(args, "/command", "option confirm off")
	//args = append(args, "/command", "option confirm off", "/log=me.log", "/loglevel=1")
	args = append(args, cmd.genOpenCmd())
	args = append(args, cmd.genstepsCmd()...)
	args = append(args, "close", "exit")
	return args
}

func (cmd *WinScpCmd) genOpenCmd() string {
	//cmdtext := "open " + cmd.user + ":" + cmd.password + "@" + cmd.host + ":" + cmd.port
	cmdtext := "open " + cmd.user + ":" + cmd.password + "@" + cmd.host + ":" + cmd.port + " -hostkey=*" + " -timeout=300"
	return cmdtext
}

func (cmd *WinScpCmd) genstepsCmd() []string {
	args := make([]string, 0)
	for _, IStep := range cmd.steps {
		args = append(args, IStep.GetStepArgs()...)
	}
	return args
}

func (cmd *WinScpCmd) SetUser(user string) {
	cmd.user = user
}

func (cmd *WinScpCmd) SetPassword(password string) {
	cmd.password = password
}

func (cmd *WinScpCmd) SetHost(host string) {
	cmd.host = host
}

func (cmd *WinScpCmd) SetPort(port string) {
	cmd.port = port
}

func (cmd *WinScpCmd) AddWinScpStep(step IWinScpStep) {
	cmd.steps = append(cmd.steps, step)
}

//// WinScpStepPutFile
type WinScpStepPutFile struct {
	localfile  string
	remotefile string
}

func NewWinScpStepPutFile() *WinScpStepPutFile {
	return &WinScpStepPutFile{}
}

func (step *WinScpStepPutFile) SetLocalfile(file string) {
	step.localfile = file
}

func (step *WinScpStepPutFile) SetRemotefile(file string) {
	step.remotefile = file
}

func (step *WinScpStepPutFile) GetStepArgs() []string {
	args := make([]string, 0)
	args = append(args, ("put " + step.localfile + " " + step.remotefile))
	return args
}

//// WinScpStepSync
type WinScpSyncDirection int

const (
	WIN_SCP_SYNC_DIRECTION_LOCAL_TO_REMOTE WinScpSyncDirection = 1
	WIN_SCP_SYNC_DIRECTION_REMOTE_TO_LOCAL WinScpSyncDirection = 2
)

type WinScpStepSync struct {
	localdir  string
	remotedir string

	// SyncDirect
	syncdirection WinScpSyncDirection

	// include file/dir, support file suffix
	include []string
	// exclude file/dir, support file suffix
	exclude []string
}

func NewWinScpStepSync() *WinScpStepSync {
	step := &WinScpStepSync{}
	step.SetDirection(WIN_SCP_SYNC_DIRECTION_LOCAL_TO_REMOTE)
	return step
}

func (step *WinScpStepSync) SetLocalDir(dir string) {
	step.localdir = dir
}

func (step *WinScpStepSync) SetRemoteDir(dir string) {
	step.remotedir = dir
}

func (step *WinScpStepSync) SetDirection(direction WinScpSyncDirection) {
	step.syncdirection = direction
}

func (step *WinScpStepSync) GetStepArgs() []string {
	args := make([]string, 0)

	// include
	if len(step.include) > 0 {
		option := ""
		for _, include := range step.include {
			if len(option) > 0 {
				option += (";" + include)
			} else {
				option += ("option include " + include)
			}
		}
		args = append(args, option)
	}

	// exclude
	if len(step.exclude) > 0 {
		option := ""
		for _, exclude := range step.exclude {
			if len(option) > 0 {
				option += (";" + exclude)
			} else {
				option += ("option exclude " + exclude)
			}
		}
		args = append(args, option)
	}

	// sync
	syncstr := "synchronize "
	if step.syncdirection == WIN_SCP_SYNC_DIRECTION_LOCAL_TO_REMOTE {
		syncstr += "remote "
	} else if step.syncdirection == WIN_SCP_SYNC_DIRECTION_REMOTE_TO_LOCAL {
		syncstr += "local "
	}
	syncstr += (step.localdir + " " + step.remotedir)
	args = append(args, syncstr)

	return args
}

func (step *WinScpStepSync) AddInclude(include string) {
	step.include = append(step.include, include)
}

func (step *WinScpStepSync) AddExclude(exclude string) {
	step.exclude = append(step.exclude, exclude)
}

//// WinScpStepCall
type WinScpStepCall struct {
	// linux shell 命令
	shellcmd []string
}

func NewWinScpStepCall() *WinScpStepCall {
	return &WinScpStepCall{}
}

func (step *WinScpStepCall) GetStepArgs() []string {
	args := make([]string, 0)
	cmdtext := ""
	for _, shellcmd := range step.shellcmd {
		if len(cmdtext) > 0 {
			cmdtext += ("; " + shellcmd)
		} else {
			cmdtext += ("call " + shellcmd)
		}
	}
	args = append(args, cmdtext)
	return args
}

func (step *WinScpStepCall) AddShellCmd(shellcmd string) {
	step.shellcmd = append(step.shellcmd, shellcmd)
}
