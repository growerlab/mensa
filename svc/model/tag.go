package model

import (
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Tag struct {
	Name   string  `json:"name"`
	Commit *Commit `json:"commit"`

	RawTag *object.Tag
}

func InitTag(name string, rawTag *object.Tag) *Tag {
	return &Tag{Name: name, RawTag: rawTag}
}
