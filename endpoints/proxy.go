package endpoints

import (
	"github.com/gin-gonic/gin"
	"log"
)

func Proxy(p ProviderData) func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		pType := c.Param("type")
		version := c.Param("version")
		os := c.Param("os")
		arch := c.Param("arch")
		filename := c.Param("filename")

		log.Printf("proxy data with namespace=%s type=%s version=%s, os=%s arch=%s filename=%s", namespace, pType, version, os, arch, filename)

		downloadData, err := p.Proxy(namespace, pType, version, os, arch, filename)
		if err != nil {
			log.Printf("INFO: error proxying data: %v\n", err)
			c.String(500, "")
			return
		}

		c.DataFromReader(200, downloadData.ContentLength, downloadData.ContentType, downloadData.Body, nil)
	}
}
