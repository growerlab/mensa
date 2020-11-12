package model

import (
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Tree struct {
	Path       string       `json:"path"`
	Name       string       `json:"name"`
	Entries    []*Entry     `json:"entries"`
	Trees      []*Tree      `json:"trees"`
	Blobs      []*Blob      `json:"blobs"`
	Submodules []*Submodule `json:"submodules"`

	RawTree *object.Tree
}

func InitTree(rawTree *object.Tree) *Tree {
	tree := &Tree{RawTree: rawTree}
	// set Path, Name...
	return tree
}
