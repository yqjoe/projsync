package cmd

// XcopyCmd is xcopy cmd
type XcopyCmd struct {
	// src files
	src string
	// dst dir
	dst string
}

func NewXcopyCmd() *XcopyCmd {
	return &XcopyCmd{"", ""}
}

func (cmd *XcopyCmd) GetCmdName() string {
	return "xcopy"
}

func (cmd *XcopyCmd) GetCmdArgs() []string {
	args := make([]string, 0)
	args = append(args, cmd.src)
	args = append(args, cmd.dst)
	args = append(args, "/s/h/y")
	return args
}

func (cmd *XcopyCmd) SetSrcFiles(src string) {
	cmd.src = src
}

func (cmd *XcopyCmd) SetDstDir(dst string) {
	cmd.dst = dst
}