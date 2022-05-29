package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/mdreem/s3_terraform_registry/logger"
)

func GetDownloadData(p ProviderData) func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		pType := c.Param("type")
		version := c.Param("version")
		os := c.Param("os")
		arch := c.Param("arch")

		logger.Sugar.Info("called get download data", "namespace", namespace, "type", pType, "version", version, "os", os, "arch", arch)

		downloadData, err := p.GetDownloadData(namespace, pType, version, os, arch)
		if err != nil {
			logger.Sugar.Error("get download data returned error", "error", err)
			c.String(500, "")
			return
		}

		c.JSON(200, downloadData)
	}
}
