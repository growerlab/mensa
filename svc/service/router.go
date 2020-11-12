package service

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/growerlab/mensa/app/conf"
	"github.com/growerlab/mensa/svc/service/middleware"
)

func BuildEngine(config *conf.Config) *gin.Engine {
	router := engine(config)
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "ok")
	})
	graphqlGroup := router.Group("/", middleware.CtxRepoMiddleware)
	{
		graphqlGroup.POST("/graphql", GraphQLHandler())
		graphqlGroup.GET("/graphql", GraphQLHandler())
		graphqlGroup.POST("/graphiql", GraphiQLHandler())
		graphqlGroup.GET("/graphiql", GraphiQLHandler())
	}
	return router
}

func engine(config *conf.Config) *gin.Engine {
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DefaultWriter = log.Writer()
	gin.DefaultErrorWriter = log.Writer()

	router := gin.New()
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	router.Use(gin.Recovery())
	return router
}
