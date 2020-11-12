package model

import (
	"errors"

	"github.com/go-git/go-git/v5/plumbing"
)

type RefType string

const (
	RefBranch RefType = "Branch"
	RefTag    RefType = "Tag"
	RefRemote RefType = "Remote"
	RefNote   RefType = "Note"

	RefUnknown RefType = "Unknown"
)

type Ref struct {
	Name    string  `json:"name"`
	RefType RefType `json:"ref_type"`
	Commit  *Commit `json:"commit"`

	RawRef *plumbing.Reference
	Repo   *Repo
}

func InitRef(name string, rawRef *plumbing.Reference) *Ref {
	return &Ref{Name: name, RawRef: rawRef}
}

func (ref *Ref) RetrieveRefType() RefType {
	target := ref.RawRef.Target()
	switch true {
	case target.IsBranch():
		ref.RefType = RefBranch
	case target.IsTag():
		ref.RefType = RefTag
	case target.IsRemote():
		ref.RefType = RefRemote
	case target.IsNote():
		ref.RefType = RefNote
	default:
		ref.RefType = RefUnknown
	}
	return ref.RefType
}

func (ref *Ref) TargetCommit() (*Commit, error) {
	refType := ref.RawRef.Type()
	switch refType {
	case plumbing.SymbolicReference:
		refName := ref.RawRef.Target()
		reference, err := ref.Repo.RawRepo.Reference(refName, false)
		if err != nil {
			return nil, err
		}
		refWrapped := &Ref{Name: refName.String(), RawRef: reference}
		return refWrapped.TargetCommit()

	case plumbing.HashReference:
		rawCommit, err := ref.Repo.RawRepo.CommitObject(ref.RawRef.Hash())
		if err != nil {
			return nil, err
		}
		return InitCommit(rawCommit), nil
	default:
		return nil, errors.New("not found target commit")
	}
}
