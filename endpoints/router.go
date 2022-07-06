package endpoints

import (
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/mdreem/s3_terraform_registry/logger"
	"time"
)

func SetupRouter(cache Cache, providerData ProviderData) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(ginzap.Ginzap(logger.Logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger.Logger, true))

	r.GET("/.well-known/terraform.json", Discovery())

	r.GET("/v1/providers/:namespace/:type/versions", ListVersions(&cache))
	r.GET("/v1/providers/:namespace/:type/:version/download/:os/:arch", GetDownloadData(providerData))

	r.GET("/proxy/:namespace/:type/:version/:filename", Proxy(providerData))
	r.GET("/refresh", RefreshHandler(&cache))
	return r
}
