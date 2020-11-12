package client

import (
	"testing"
)

func TestBranchInfo(t *testing.T) {
	client, repo := defaultClient()
	branch := client.Branch(repo)
	branch.client = &DirectGQLClient{}

	defaultBranch, branches, err := branch.Info()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	if defaultBranch != "master" {
		t.Fail()
	}

	if len(branches) != len([]string{"master"}) {
		t.Fatal(branches)
	}
}
