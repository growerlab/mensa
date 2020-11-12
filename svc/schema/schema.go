package schema

import (
	"github.com/graphql-go/graphql"
)

var rootQuery = func() *graphql.Object {
	fields := map[string]*graphql.Field{
		queryRepo.Name: &queryRepo,
	}
	return graphql.NewObject(graphql.ObjectConfig{
		Name:        "RootQuery",
		Description: "Root Query",
		Fields:      graphql.Fields(fields),
	})
}()

var rootMutation = func() *graphql.Object {
	fields := map[string]*graphql.Field{
		createRepo.Name:   &createRepo,
		deleteBranch.Name: &deleteBranch,
	}
	return graphql.NewObject(graphql.ObjectConfig{
		Name:        "RootMutation",
		Description: "Root mutation",
		Fields:      graphql.Fields(fields),
	})
}()

var Schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    rootQuery,
	Mutation: rootMutation,
})
