package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/mdreem/s3_terraform_registry/cache"
	"github.com/mdreem/s3_terraform_registry/logger"
)

func refreshHandler(cache *cache.Cache) func(c *gin.Context) {
	return func(c *gin.Context) {
		logger.Sugar.Infow("refreshing cache")

		err := (*cache).Refresh()
		if err != nil {
			logger.Sugar.Errorw("error refreshing data", "error", err)
			c.String(500, "")
			return
		}

		c.String(200, "refreshed")
	}
}
