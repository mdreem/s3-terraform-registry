package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/mdreem/s3_terraform_registry/schema"
)

func discovery() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(200, schema.Discovery{ProvidersV1: "/v1/providers/"})
	}
}
