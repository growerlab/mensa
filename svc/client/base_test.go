package client

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/growerlab/mensa/svc/schema"
	"github.com/growerlab/mensa/svc/service/middleware"
)

func init() {
	base := filepath.Join(os.Getenv("GOPATH"), "src", "github.com/growerlab/mensa/svc")
	_ = os.Chdir(base)
}

func TestPost(t *testing.T) {
}

func defaultClient() (*Client, *RepoContext) {
	client, _ := NewClient("", 0)
	repo := &RepoContext{
		Path: "/",
		Name: "moli",
	}
	return client, repo
}

type DirectGQLClient struct {
}

func (f *DirectGQLClient) Query(req *Request) (*Result, error) {
	body, _ := json.Marshal(req.RequestBody())
	gqlResult, err := graphqlExecuter(body)
	if err != nil {
		return nil, err
	}
	gqlResultData, err := json.Marshal(gqlResult)
	if err != nil {
		return nil, err
	}
	return BuildResult(gqlResultData)
}

func (f *DirectGQLClient) Mutation(req *Request) (*Result, error) {
	return f.Query(req)
}

func graphqlExecuter(body []byte) (result *graphql.Result, err error) {
	var opts handler.RequestOptions
	err = json.Unmarshal(body, &opts)
	if err != nil {
		return nil, err
	}

	// execute graphql query
	ctx := context.Background()
	ctx, _ = middleware.BuildRepoContext(ctx, opts.Variables)

	params := graphql.Params{
		Schema:         schema.Schema,
		RequestString:  opts.Query,
		VariableValues: opts.Variables,
		OperationName:  opts.OperationName,
		Context:        ctx,
	}
	result = graphql.Do(params)
	return
}
