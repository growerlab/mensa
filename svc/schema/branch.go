package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/pkg/errors"
)

var branchType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Repo",
	Description: "Repo Model",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"ref": &graphql.Field{
			Type: graphql.String,
		},
		"commits": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// mutationDeleteBranch delete branch
var deleteBranch = graphql.Field{
	Name:        "deleteBranch",
	Description: "delete branch",
	Type:        graphql.Boolean,
	Args: graphql.FieldConfigArgument{
		"branchName": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
	Resolve: func(p graphql.ResolveParams) (result interface{}, err error) {
		result = false
		branchName, ok := p.Args["branchName"].(string)
		if !ok {
			err = errors.New("branch name is required")
			return
		}
		repo, err := loadRepo(&p)
		if err != nil {
			return
		}
		err = repo.DeleteBranch(branchName)
		if err != nil {
			return
		}
		result = true
		return
	},
}
