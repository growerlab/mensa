package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/growerlab/mensa/svc/model"
	"github.com/pkg/errors"
)

var RepoType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Repo",
	Description: "Repo Model",
	Fields: graphql.Fields{
		"path": &graphql.Field{
			Type: graphql.String,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"repo_size": &graphql.Field{
			Type: graphql.Float,
		},
		"default_branch": &graphql.Field{
			Type: branchType,
		},
		"branches": &graphql.Field{
			Type:        graphql.NewList(branchType),
			Description: "branch list",
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				repo := params.Context.Value("repo").(*model.Repo)
				return repo.Branches, nil
			},
		},
	},
})

var queryRepo = graphql.Field{
	Name:        "repo",
	Description: "Query repo",
	Type:        graphql.NewNonNull(RepoType),
	Resolve: func(p graphql.ResolveParams) (result interface{}, err error) {
		result, err = loadRepo(&p)
		return
	},
}

var createRepo = graphql.Field{
	Name:        "createRepo",
	Description: "Create Repo",
	Type:        graphql.NewNonNull(RepoType),
	Args: graphql.FieldConfigArgument{
		"path": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"name": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
	Resolve: func(p graphql.ResolveParams) (result interface{}, err error) {
		path, _ := p.Args["path"].(string)
		name, _ := p.Args["name"].(string)
		repo, err := model.InitRepo(path, name)
		if err != nil {
			return nil, err
		}
		return repo, nil
	},
}

func loadRepo(p *graphql.ResolveParams) (*model.Repo, error) {
	repo, ok := p.Context.Value("repo").(*model.Repo)
	if !ok {
		return nil, errors.New("repo is required")
	}
	return repo, nil
}
