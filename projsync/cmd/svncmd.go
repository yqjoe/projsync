package cmd

// SvnCmd is cmd of svn
type SvnCmd struct {
	// proj svn dir
	svndir string

	// opration
	op string

	// username
	username string
	
	// password
	password string
}

func NewSvnCmd() *SvnCmd {
	return &SvnCmd{}
}

func (cmd *SvnCmd) GetCmdName() string {
	return "svn"
}

func (cmd *SvnCmd) GetCmdArgs() []string {
	args := make([]string, 0)
	args = append(args, cmd.op)
	args = append(args, cmd.svndir)
	args = append(args, "--username", cmd.username, "--password", cmd.password)
	return args
}

func (cmd *SvnCmd) SetOp(op string) {
	cmd.op = op
}

func (cmd *SvnCmd) SetSvnDir(dir string) {
	cmd.svndir = dir
}

func (cmd *SvnCmd) SetUser(username string) {
	cmd.username = username
}

func (cmd *SvnCmd) SetPassword(password string) {
	cmd.password = password
}