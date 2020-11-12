package model

import (
	"os"
	"path/filepath"
	"testing"
)

func init() {
	base := filepath.Join(os.Getenv("GOPATH"), "src", "github.com/growerlab/mensa/svc")
	_ = os.Chdir(base)

	ReposDir = filepath.Join(base, "repos")

	err := initRepo("", "test")
	if err != nil {
		panic(err)
	}
}

func TestInitRepo(t *testing.T) {
	if _, err := os.Stat(ReposDir + "/test/HEAD"); os.IsNotExist(err) {
		t.Errorf("init repo faild: %+v", err)
	}
	if _, err := os.Stat(ReposDir + "/test/hooks"); os.IsNotExist(err) {
		t.Errorf("init repo hooks faild: %+v", err)
	}
}

func initRepo(repoPath, name string) error {
	var err error
	_, err = InitRepo(repoPath, name)
	if err != nil {
		return err
	}

	return nil
}
