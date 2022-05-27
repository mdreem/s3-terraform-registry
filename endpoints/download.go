package endpoints

import (
	"github.com/gin-gonic/gin"
	"log"
)

func GetDownloadData(p ProviderData) func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		pType := c.Param("type")
		version := c.Param("version")
		os := c.Param("os")
		arch := c.Param("arch")

		log.Printf("called get download data with namespace=%s type=%s version=%s, os=%s arch=%s", namespace, pType, version, os, arch)

		downloadData, err := p.GetDownloadData(namespace, pType, version, os, arch)
		if err != nil {
			log.Printf("ERROR: get download data returned: %v\n", err)
			c.String(500, "")
			return
		}

		c.JSON(200, downloadData)
	}
}
