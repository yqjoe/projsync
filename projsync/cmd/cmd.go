package cmd

import (
	//"bytes"
	"fmt"
	"os/exec"
	"io"
)

// ICmd interface
type ICmd interface {
	GetCmdName() string
	GetCmdArgs() []string
}

// ExecCmd is function of exec windows batch cmd
func ExecCmd(cmd ICmd, printer io.Writer) {
	execCmd(printer, cmd.GetCmdName(), cmd.GetCmdArgs()...)
}

func execCmd(printer io.Writer, cmdname string, args ...string) {
	//fmt.Println(args)
	//for _, arg := range(args) {
	//	fmt.Println(arg)
	//}

	c := exec.Command(cmdname, args...)
	//var cmderr bytes.Buffer
	c.Stdout = printer 
	//c.Stderr = &cmderr

	if err := c.Run(); err != nil {
		fmt.Printf("Error:%v cmd:%v\n", err, cmdname)
		//fmt.Println("ErrInfo:", cmderr.String())
	}
}