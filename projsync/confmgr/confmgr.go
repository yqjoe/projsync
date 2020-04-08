package confmgr

import (
	//"time"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/robfig/config"
)

type StepConf struct {
	StepName          string
	Relativedir       string
	DstRelativedir    string
	OverrideLocaldir  string
	OverrideRemotedir string
	FileName          string
	SyncDirection     string
	Include           []string `xml:"IncludeList>Include"`
	Exclude           []string `xml:"ExcludeList>Exclude"`
	ShellCmd          []string `xml:"ShellCmdList>ShellCmd"`
}

type CmdConf struct {
	CmdName string
	Step    []StepConf `xml:"StepList>Step"`
}

type TaskConf struct {
	TaskName         string
	TaskPrinter      string
	AutoDoTask       bool
	AutoDoTaskCircle int
	Cmd              []CmdConf `xml:"CmdList>Cmd"`
}

type ProjConf struct {
	Projectname string
	User        string
	Password    string
	Host        string
	Port        string
	Localdir    string
	Remotedir   string
	SvnUser     string
	SvnPassword string
	Task        []TaskConf `xml:"TaskList>Task"`
}

type ProjConfMap map[string]*ProjConf

type ConfMgr struct {
	ProjectCnt     int
	TaskServerPort int
	ProjectMap     ProjConfMap
}

const (
	CFG_FILE_NAME = "projsync.ini"
)

func readString(conf *config.Config, sec string, key string) (string, error) {
	str, err := conf.String(sec, key)
	if err != nil {
		fmt.Println("Read fail Sec:", sec, " Key:", key, " Err:", err)
		return "", err
	}

	return str, nil
}

func readInt(conf *config.Config, sec string, key string) (int, error) {
	num, err := conf.Int(sec, key)
	if err != nil {
		fmt.Println("Read fail Sec:", sec, " Key:", key, " Err:", err)
		return 0, err
	}

	return num, nil
}

func openXml(xmlfile string) ([]byte, error) {
	fd, open_err := os.Open(xmlfile)
	if open_err != nil {
		return nil, open_err
	}
	defer fd.Close()

	data, read_err := ioutil.ReadAll(fd)
	if read_err != nil {
		return nil, read_err
	}

	return data, nil
}

var ConfObj ConfMgr

func Init() error {
	c, err := config.ReadDefault(CFG_FILE_NAME)
	if err != nil {
		fmt.Println("Read CFGFile fail:", CFG_FILE_NAME)
		return err
	}
	ConfObj = ConfMgr{0, 0, make(ProjConfMap, 0)}

	ConfObj.ProjectCnt, err = readInt(c, "base", "projectcnt")
	if err != nil {
		return err
	}

	ConfObj.TaskServerPort, err = readInt(c, "base", "taskserverport")
	if err != nil {
		return err
	}

	for i := 1; i <= ConfObj.ProjectCnt; i++ {
		key := "project" + strconv.Itoa(i)
		projectname, err := readString(c, "base", key)
		if err != nil {
			return err
		}

		ProjectConfObj := ProjConf{}
		data, open_err := openXml(projectname + ".xml")
		if open_err != nil {
			fmt.Println("openxml fail:", projectname)
			return open_err
		}

		um_err := xml.Unmarshal(data, &ProjectConfObj)
		if um_err != nil {
			fmt.Println("unmarshal fail:", projectname)
			return um_err
		}

		ConfObj.ProjectMap[ProjectConfObj.Projectname] = &ProjectConfObj

		formatConf(&ProjectConfObj)
	}

	return nil
}

func GetProjectConf(projectname string) *ProjConf {
	projectconf, ok := ConfObj.ProjectMap[projectname]
	if ok == false {
		return nil
	}

	return projectconf
}

func GetTaskConf(projectname, taskname string) *TaskConf {
	projectconf := GetProjectConf(projectname)
	if projectconf == nil {
		return nil
	}

	for _, task := range projectconf.Task {
		if task.TaskName == taskname {
			return &task
		}
	}

	return nil
}

func GetTaskServerAddr() string {
	return (":" + strconv.Itoa(ConfObj.TaskServerPort))
}

func formatConf(projconf *ProjConf) {
	recurseRftObj(reflect.ValueOf(projconf), "${Projectname}", projconf.Projectname)
	recurseRftObj(reflect.ValueOf(projconf), "${User}", projconf.User)
	recurseRftObj(reflect.ValueOf(projconf), "${Remotedir}", projconf.Remotedir)
	recurseRftObj(reflect.ValueOf(projconf), "${Localdir}", projconf.Localdir)
}

func recurseRftObj(value reflect.Value, srcstr, deststr string) {
	switch value.Kind() {
	case reflect.Ptr:
		recurseRftObj(value.Elem(), srcstr, deststr)
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			recurseRftObj(value.Field(i), srcstr, deststr)
		}
	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			recurseRftObj(value.Index(i), srcstr, deststr)
		}
	case reflect.String:
		value.SetString(replaceStr(value.String(), srcstr, deststr))
	}
}

func replaceStr(str, srcstr, deststr string) string {
	str = strings.Replace(str, srcstr, deststr, -1)
	return str
}
