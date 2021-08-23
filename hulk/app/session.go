package app

import (
	"encoding/json"
	"os"

	"github.com/growerlab/mensa/hulk/repo"
	"github.com/pkg/errors"
)

type Action string

const (
	ActionCreated Action = "created" // create branch or tag
	ActionRemoved Action = "removed" // remove branch or tag
	ActionPushed  Action = "pushed"  // push commit
)

type RefType string

const (
	RefTypeBranch RefType = "branch"
	RefTypeTag    RefType = "tag"
)

type PushSession struct {
	After  string `json:"after"`  // old
	Before string `json:"before"` // new
	Ref    string `json:"ref"`    // branch, tag name

	RepoDir string `json:"repo_dir"`

	RepoOwner string `json:"repo_owner"` // namespace.path
	RepoPath  string `json:"repo_path"`  // repository name

	Action  Action  `json:"action"`
	RefType RefType `json:"ref_type"`

	ProtType string `json:"prot_type"` // http/ssh

	Operator string `json:"operator"` // 推送者
}

func (r *PushSession) JSON() string {
	if r == nil {
		return ""
	}
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *PushSession) IsNullOldCommit() bool {
	return r.After == repo.ZeroRef
}

func (r *PushSession) IsNullNewCommit() bool {
	return r.Before == repo.ZeroRef
}

func (r *PushSession) IsNewBranch() bool {
	return repo.IsBranch(r.Ref) && r.IsNullOldCommit() && !r.IsNullNewCommit()
}

func (r *PushSession) IsNewTag() bool {
	return repo.IsTag(r.Ref) && r.IsNullOldCommit() && !r.IsNullNewCommit()
}

func (r *PushSession) IsCommitPush() bool {
	return !r.IsNullOldCommit() && !r.IsNullNewCommit()
}

func (r *PushSession) prepare() error {
	r.RepoOwner = RepoOwner
	r.RepoPath = RepoPath

	if r.IsCommitPush() {
		r.RefType = RefTypeBranch
		r.Action = ActionPushed
	} else {
		if repo.IsBranch(r.Ref) {
			r.RefType = RefTypeBranch
			if r.IsNewBranch() {
				r.Action = ActionCreated
			} else {
				r.Action = ActionRemoved
			}
		} else if repo.IsTag(r.Ref) {
			r.RefType = RefTypeTag
			if r.IsNewTag() {
				r.Action = ActionCreated
			} else {
				r.Action = ActionRemoved
			}
		} else {
			return errors.Errorf("invalid ref '%s'", r.Ref)
		}
	}
	return nil
}

func Session() *PushSession {
	pwd, err := os.Getwd()
	if err != nil {
		ErrPanic(err)
	}

	ctx := &PushSession{
		RepoDir: pwd,
		After:   os.Args[1],
		Before:  os.Args[2],
		Ref:     os.Args[3],
	}

	if err := ctx.prepare(); err != nil {
		ErrPanic(err)
	}
	return ctx
}

func ErrPanic(err error) {
	if err != nil {
		panic(errors.WithStack(err))
	}
}
