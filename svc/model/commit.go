package model

import (
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Commit struct {
	Sha       string    `json:"sha"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	committer string    `json:"committer"`
	Parent    *Commit   `json:"parent"`
	Parents   []*Commit `json:"parents"`
	Tree      *Tree     `json:"tree"`

	RawCommit *object.Commit
}

func InitCommit(rawCommit *object.Commit) *Commit {
	commit := &Commit{RawCommit: rawCommit}
	// set Sha, Message...
	return commit
}
