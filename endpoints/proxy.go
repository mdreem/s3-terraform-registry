package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/mdreem/s3_terraform_registry/logger"
)

func Proxy(p ProviderData) func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		pType := c.Param("type")
		version := c.Param("version")
		os := c.Param("os")
		arch := c.Param("arch")
		filename := c.Param("filename")

		logger.Info("proxy data with", "namespace", namespace, "type", pType, "version", version, "os", os, "arch", arch, "filename", filename)

		downloadData, err := p.Proxy(namespace, pType, version, os, arch, filename)
		if err != nil {
			logger.Error("error proxying data", "error", err)
			c.String(500, "")
			return
		}

		c.DataFromReader(200, downloadData.ContentLength, downloadData.ContentType, downloadData.Body, nil)
	}
}
