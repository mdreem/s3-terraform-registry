package endpoints

import (
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/mdreem/s3_terraform_registry/logger"
	"github.com/mdreem/s3_terraform_registry/providerdata"
	"time"
)

type CacheableProviderData interface {
	providerdata.ProviderData
	Cache
}

func SetupRouter(cacheableProviderData CacheableProviderData) *gin.Engine {
	providerData, ok := cacheableProviderData.(providerdata.ProviderData)
	if !ok {
		logger.Sugar.Panicw("unable to cast to ProviderData.")
	}

	cache, ok := cacheableProviderData.(Cache)
	if !ok {
		logger.Sugar.Panicw("unable to cast to Cache.")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(ginzap.Ginzap(logger.Logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger.Logger, true))

	r.GET("/.well-known/terraform.json", discovery())

	r.GET("/v1/providers/:namespace/:type/versions", listVersions(&providerData))
	r.GET("/v1/providers/:namespace/:type/:version/download/:os/:arch", getDownloadData(&providerData))

	r.GET("/proxy/:namespace/:type/:version/:filename", proxy(&providerData))
	r.GET("/refresh", refreshHandler(&cache))

	return r
}
