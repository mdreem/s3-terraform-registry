package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/mdreem/s3_terraform_registry/logger"
)

func ListVersions(cache *Cache) func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		providertype := c.Param("type")

		logger.Sugar.Info("called list versions ", "namespace", namespace, "providertype", providertype)

		versions, err := (*cache).ListVersions(namespace, providertype)
		if err != nil {
			logger.Sugar.Error("list versions returned error", "error", err)
			c.String(500, "")
			return
		}

		c.JSON(200, versions)
	}
}
