package app

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/growerlab/mensa/app/conf"
	"github.com/pkg/errors"
)

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
	cmd.Stderr = gin.DefaultErrorWriter
	err := cmd.Run()
	return errors.WithStack(err)
}
