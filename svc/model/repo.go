package model

import (
	"os"
	"path"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/growerlab/mensa/svc/model/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var ReposDir = "repos/"

const DefaultBranch = "master"

type Repo struct {
	Path          string  `json:"path"`
	Name          string  `json:"name"`
	DefaultBranch *Branch `json:"default_branch"`

	// bytes
	RepoSize float64 `json:"repo_size"`

	Branches []*Branch `json:"branches"`

	Tags []*Tag `json:"tags"`

	Refs []*Ref `json:"refs"`

	// Submodules []*Submodule `json:"submodules"`

	// internal methods
	RawRepo  *git.Repository
	RepoPath string
}

func OpenRepo(repoPath string, name string) (*Repo, error) {
	repo := &Repo{
		Path: repoPath,
		Name: name,
	}
	log.Info().Str("path", repo.RepoPath).Msg("open repo")
	repo.RepoPath = path.Join(ReposDir, repoPath, name)
	rawRepo, err := git.PlainOpen(repo.RepoPath)
	if err != nil {
		return nil, err
	}

	repo.RawRepo = rawRepo
	repo.postRepoCreated()

	return repo, nil
}

// TODO 如果仓库被fork过，用户删除时，不应该进行物理删除，那么用户再创建同名仓库时，应该做特殊处理
//
func InitRepo(repoPath string, name string) (*Repo, error) {
	if len(repoPath) == 0 || len(name) == 0 {
		return nil, errors.New("path and name is required")
	}
	repo := &Repo{
		Path: repoPath,
		Name: name,
	}
	repo.RepoPath = path.Join(ReposDir, repoPath)

	rawRepo, err := git.PlainInit(repo.RepoPath, true)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	repo.RawRepo = rawRepo
	repo.postRepoCreated()
	return repo, nil
}

func (repo *Repo) postRepoCreated() {
	// fill all fields after repo oject created

	// References
	repo.Refs = make([]*Ref, 0)
	refsIterator, err := repo.RawRepo.References()
	if err == nil {
		_ = refsIterator.ForEach(func(rawRef *plumbing.Reference) error {
			repo.Refs = append(repo.Refs, InitRef(rawRef.Name().String(), rawRef))
			return nil
		})
	}

	// Branches
	repo.Branches, err = repo.branches()
	if err != nil {
		return
	}

	// Tags
	repo.Tags = make([]*Tag, 0)
	tagsInterator, err := repo.RawRepo.Tags()
	if err == nil {
		_ = tagsInterator.ForEach(func(tag *plumbing.Reference) error {
			rawTag, _ := repo.RawRepo.TagObject(tag.Hash())
			if rawTag != nil {
				repo.Tags = append(repo.Tags, InitTag(tag.Name().String(), rawTag))
			}
			return nil
		})
	}

	// Submodules
	// repo.Submodules = make([]*Submodule, 0)
	// tree, err := repo.RawRepo.Worktree()
	// if err == nil {
	// 	submodules, err := tree.Submodules()
	// 	if err == nil {
	// 		for _, sub := range submodules {
	// 			repo.Submodules = append(repo.Submodules, InitSubmodule(sub))
	// 		}
	// 	}
	// }
}

func (repo *Repo) Head() (*Ref, error) {
	rawRef, err := repo.RawRepo.Head()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ref := &Ref{Name: rawRef.Name().String(), RawRef: rawRef}

	return ref, nil
}

func (repo *Repo) FileEntries(path string, hash plumbing.Hash) ([]object.TreeEntry, error) {
	tree, err := repo.RawRepo.TreeObject(hash)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pathTree, err := tree.Tree(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return pathTree.Entries, nil
}

func (repo *Repo) branches() ([]*Branch, error) {
	if len(repo.Branches) > 0 {
		return repo.Branches, nil
	}

	refHead, err := repo.RawRepo.Head()
	if err != nil {
		return nil, err
	}

	iter, err := repo.RawRepo.Branches()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	branches := make([]*Branch, 0)
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		branch := InitBranch(ref)
		if utils.ReferenceCompare(ref, refHead) {
			branch.Default = true
			repo.DefaultBranch = branch
		}
		branches = append(branches, branch)
		return nil
	})
	repo.Branches = branches
	return repo.Branches, errors.WithStack(err)
}

func (repo *Repo) CreateBranch(name string) error {
	err := repo.RawRepo.CreateBranch(&config.Branch{
		Name: name,
	})
	return errors.WithStack(err)
}

func (repo *Repo) DeleteBranch(name string) error {
	err := repo.RawRepo.DeleteBranch(name)
	repo.postRepoCreated()
	return errors.WithStack(err)
}

func (repo *Repo) Size() int64 {
	var size int64
	err := filepath.Walk(repo.RepoPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	if err != nil {
		return 0
	}
	return size
}
