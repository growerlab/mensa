package model

import "github.com/go-git/go-git/v5"

type Submodule struct {
	RawSubmodule *git.Submodule
}

func InitSubmodule(rawSubmodule *git.Submodule) *Submodule {
	submodule := &Submodule{RawSubmodule: rawSubmodule}
	return submodule
}
