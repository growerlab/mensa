package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/growerlab/mensa/app/conf"
	"github.com/pkg/errors"
)

type Option struct {
	Name string
	Args string
}

var GitReceivePackOptions = []*Option{
	{"-c", fmt.Sprintf("core.hooksPath=%s", filepath.Join(os.Args[0], "hooks"))},
	{"-c", "core.alternateRefsCommand=exit 0 #"},
	{"-c", "receive.fsck.badTimezone=ignore"},
}

func gitCommand(in io.Reader, out io.Writer, repoDir string, args []string, envSet map[string]string) error {

	gitBinPath := conf.GetConfig().GitPath
	deadline := time.Duration(conf.GetConfig().Deadline) * time.Second

	// deadline
	cmdCtx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, gitBinPath, args...)
	if len(envSet) > 0 {
		cmd.Env = make([]string, 0, len(envSet))
		for k, v := range envSet {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	cmd.Dir = repoDir
	if in != nil {
		cmd.Stdin = in
	}
	if out != nil {
		cmd.Stdout = out
	}
	cmd.Stderr = out
	err := cmd.Run()
	return errors.WithStack(err)
}
