package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/mdreem/s3_terraform_registry/logger"
)

func ListVersions(providerData *ProviderData) func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		providerType := c.Param("type")

		logger.Sugar.Infow("called list versions ", "namespace", namespace, "providerType", providerType)

		versions, err := (*providerData).ListVersions(namespace, providerType)
		if err != nil {
			logger.Sugar.Errorw("list versions returned error", "error", err)
			c.String(500, "")
			return
		}

		c.JSON(200, versions)
	}
}
