package service

import (
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/handler"
	"github.com/growerlab/mensa/svc/schema"
)

func GraphQLHandler() gin.HandlerFunc {
	h := handler.New(&handler.Config{
		Schema: &schema.Schema,
		Pretty: true,
	})

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func GraphiQLHandler() gin.HandlerFunc {
	h := handler.New(&handler.Config{
		Schema:   &schema.Schema,
		Pretty:   true,
		GraphiQL: true,
	})

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
