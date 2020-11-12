package utils

import "github.com/go-git/go-git/v5/plumbing"

func ReferenceCompare(a, b *plumbing.Reference) bool {
	if a != nil && b != nil {
		return a.Hash() == b.Hash()
	} else if a == nil && b == nil {
		return true
	}
	return false
}
