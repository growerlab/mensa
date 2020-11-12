package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/growerlab/mensa/svc/model"
)

func TestRepository_Create(t *testing.T) {
	repository := &Repository{
		client: &DirectGQLClient{},
		repo: &RepoContext{
			Path: "/",
			Name: "moli33",
		},
	}

	err := repository.Create()
	if err != nil {
		t.Errorf("repository create err: %v", err)
	}

	repoDir := filepath.Join(model.ReposDir, repository.repo.Path, repository.repo.Name)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		t.Error(err)
	} else {
		// clearn？
	}
}
