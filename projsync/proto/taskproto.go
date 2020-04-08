package proto

// 加任务协议
type ReqAddTask struct {
	ProjectName string
	TaskName    string

	Putstepfile string // winscp put 命令 同步的文件，本地文件的全路径

	CallParams []string // winscp call 命令的参数，如果参数不为空，依次填到这里

	// 同步工具打印日志服务器的地址，如果不开启，则为空
	SyncToolPrintSvrAddr string
}

type RspAddTask struct {
	Ret int
	Err string
}

// 输出任务执行信息协议
type ReqPrintTaskInfo struct {
	Info string
}

type RspPrintTaskInfo int

// 关闭printersvr
type ReqClosePrinterSvr int
type RspClosePrinterSvr int

type ReqGetAutoClose struct {
	ProjectName string
}

type RspGetAutoClose struct {
	ProjectName string
	AutoClose   bool
}

type ReqSetAutoClose struct {
	ProjectName string
	AutoClose   bool
}

type RspSetAutoClose struct {
	Ret int
	Err string
}

type ReqAutoInfo struct {
	ProjectName string
}

type AutoInfo struct {
	TaskName        string
	AutoCircle      int
	LastDoTimestamp int64
}

type RspAutoInfo struct {
	ProjectName   string
	AutoInfoArray []AutoInfo
	AutoClose     bool
}
