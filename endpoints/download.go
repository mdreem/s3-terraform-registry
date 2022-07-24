package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/mdreem/s3_terraform_registry/logger"
)

func GetDownloadData(providerData *ProviderData) func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		providerType := c.Param("type")
		version := c.Param("version")
		os := c.Param("os")
		arch := c.Param("arch")

		logger.Sugar.Infow("called get download data", "namespace", namespace, "type", providerType, "version", version, "os", os, "arch", arch)

		downloadData, err := (*providerData).GetDownloadData(namespace, providerType, version, os, arch)
		if err != nil {
			logger.Sugar.Errorw("get download data returned error", "error", err)
			c.String(500, "")
			return
		}

		c.JSON(200, downloadData)
	}
}
