package middleware

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/handler"
	"github.com/growerlab/mensa/svc/model"
	"github.com/pkg/errors"
)

func CtxRepoMiddleware(c *gin.Context) {
	if c.Request.URL.Path == "/graphql" {
		bodyRaw, err := c.GetRawData()
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, errors.WithStack(err))
			return
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyRaw))
		reqOptions := handler.NewRequestOptions(c.Request)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyRaw))

		repoCtx, err := BuildRepoContext(c, reqOptions.Variables)
		if err == nil {
			c.Request = c.Request.WithContext(repoCtx)
		} else {
			log.Println("not found repo: ", reqOptions.Variables)
		}
	}
	c.Next()
}

type RepoRequest struct {
	Path string
	Name string
}

func (r *RepoRequest) fullPath() string {
	return filepath.Join(r.Path, r.Name)
}

func BuildRepoContext(c context.Context, varsMap map[string]interface{}) (context.Context, error) {
	reqRepo, err := getRepo(varsMap)
	if err != nil {
		return nil, err
	}
	repo, err := model.OpenRepo(reqRepo.Path, reqRepo.Name)
	if err != nil {
		return nil, err
	}
	return context.WithValue(c, "repo", repo), nil
}

func getRepo(variables map[string]interface{}) (*RepoRequest, error) {
	var repo = RepoRequest{
		Path: variables["path"].(string),
		Name: variables["name"].(string),
	}

	if len(repo.Path) == 0 || len(repo.Name) == 0 {
		return nil, errors.New("not found repo.Path or repo.Name")
	}

	return &repo, nil
}
