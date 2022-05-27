package endpoints

import (
	"github.com/gin-gonic/gin"
	"log"
)

func ListVersions(cache *Cache) func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		providertype := c.Param("type")

		log.Printf("called list versions with namespace=%s type=%s", namespace, providertype)

		versions, err := (*cache).ListVersions(namespace, providertype)
		if err != nil {
			log.Printf("ERROR: list versions returned: %v\n", err)
			c.String(500, "")
			return
		}

		c.JSON(200, versions)
	}
}
