package app

import (
	"math"

	"github.com/growerlab/mensa/hulk/repo"
	"github.com/pkg/errors"
)

var _ Hook = (*HookEvent)(nil)

type PushEvent struct {
	*PushSession
	CommitCount int    `json:"commit_count"`
	RefCount    int    `json:"ref_count"`
	Message     string `json:"commit_message"` // commit/tag message
}

// 创建推送事件
type HookEvent struct {
}

func (h *HookEvent) Label() string {
	return "event"
}

func (h *HookEvent) Priority() uint {
	return math.MaxUint32
}

func (h *HookEvent) Process(sess *PushSession) error {
	var repository = repo.NewRepository(sess.RepoDir)
	var event *PushEvent
	var err error

	if sess.IsNewTag() {
		event, err = h.buildNewTagEvent(repository, sess)
	} else if sess.IsNewBranch() {
		event, err = h.buildNewBranchEvent(repository, sess)
	} else if sess.IsCommitPush() {
		event, err = h.buildCommitEvent(repository, sess)
	} else {
		return errors.Errorf("invalid push session: '%s'", sess.JSON())
	}
	if err != nil {
		return errors.WithStack(err)
	}
	return h.push(event)
}

// 将event推送给redis的stream
// TODO backend项目响应这个事件
func (h *HookEvent) push(event *PushEvent) error {
	return nil
}

func (h *HookEvent) buildCommitEvent(repository *repo.Repository, sess *PushSession) (*PushEvent, error) {
	commits, err := repository.BetweenCommits(sess.Before, sess.After, 20)
	if err != nil {
		return nil, err
	}

	return &PushEvent{
		PushSession: sess,
		CommitCount: len(commits),
		RefCount:    1,
		Message:     "",
	}, nil
}

func (h *HookEvent) buildNewBranchEvent(repository *repo.Repository, sess *PushSession) (*PushEvent, error) {
	_, err := repository.BranchByRef(sess.Ref)
	if err != nil {
		return nil, err
	}

	return &PushEvent{
		PushSession: sess,
		CommitCount: 0,
		RefCount:    1,
		Message:     "",
	}, nil
}

func (h *HookEvent) buildNewTagEvent(repository *repo.Repository, sess *PushSession) (*PushEvent, error) {
	tag, err := repository.TagByHash(sess.Before)
	if err != nil {
		return nil, err
	}

	return &PushEvent{
		PushSession: sess,
		CommitCount: 0,
		RefCount:    1,
		Message:     tag.Message,
	}, nil
}

type Message struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
}
