package app

import (
	"math"

	"github.com/growerlab/mensa/hulk/repo"
	"github.com/pkg/errors"
)

var _ Hook = (*HookEvent)(nil)

const (
	MaxCommitLimit = 20
)

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

func (h *HookEvent) Process(dispatcher EventDispatcher, sess *PushSession) error {
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
		return errors.Errorf("invalid session: '%s'", sess.JSON())
	}
	if err != nil {
		return errors.WithStack(err)
	}
	return dispatcher.Dispatch(event)
}

func (h *HookEvent) buildCommitEvent(repository *repo.Repository, sess *PushSession) (*PushEvent, error) {
	commits, err := repository.BetweenCommits(sess.Before, sess.After, MaxCommitLimit)
	if err != nil {
		return nil, err
	}

	plainCommits := repo.BuildPlainCommits(commits...)
	message, err := plainCommits.ToString()
	if err != nil {
		return nil, err
	}

	return &PushEvent{
		PushSession: sess,
		CommitCount: len(commits),
		RefCount:    1,
		Message:     message,
	}, nil
}

func (h *HookEvent) buildNewBranchEvent(repository *repo.Repository, sess *PushSession) (*PushEvent, error) {
	_, err := repository.BranchByRef(sess.Ref)
	if err != nil {
		return nil, err
	}

	commits, err := repository.BetweenCommits(sess.Before, sess.After, MaxCommitLimit)
	if err != nil {
		return nil, err
	}

	plainCommits := repo.BuildPlainCommits(commits...)
	message, err := plainCommits.ToString()
	if err != nil {
		return nil, err
	}

	return &PushEvent{
		PushSession: sess,
		CommitCount: len(commits),
		RefCount:    1,
		Message:     message,
	}, nil
}

func (h *HookEvent) buildNewTagEvent(repository *repo.Repository, sess *PushSession) (*PushEvent, error) {
	tag, err := repository.TagByHash(sess.Before)
	if err != nil {
		return nil, err
	}

	commits, err := repository.BetweenCommits(sess.Before, sess.After, MaxCommitLimit)
	if err != nil {
		return nil, err
	}

	return &PushEvent{
		PushSession: sess,
		CommitCount: len(commits),
		RefCount:    1,
		Message:     tag.Message,
	}, nil
}
