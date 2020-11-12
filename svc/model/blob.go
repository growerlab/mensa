package model

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	// git "github.com/go-git/go-git/v5
)

type Blob struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Content string `json:"content"`

	RawBlob *object.Blob
}

func InitBlob(rawBlob *object.Blob) *Blob {
	blob := &Blob{RawBlob: rawBlob}
	// set Path, Name...
	return blob
}
