package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/mdreem/s3_terraform_registry/logger"
	"github.com/mdreem/s3_terraform_registry/providerdata"
)

func proxy(providerData *providerdata.ProviderData) func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		providerType := c.Param("type")
		version := c.Param("version")
		filename := c.Param("filename")

		logger.Sugar.Infow("proxy data with", "namespace", namespace, "type", providerType, "version", version, "filename", filename)

		downloadData, err := (*providerData).Proxy(namespace, providerType, version, filename)
		if err != nil {
			logger.Sugar.Errorw("error proxying data", "error", err)
			c.String(500, "")
			return
		}

		c.DataFromReader(200, downloadData.ContentLength, downloadData.ContentType, downloadData.Body, nil)
	}
}
