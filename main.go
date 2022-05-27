package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mdreem/s3_terraform_registry/endpoints"
	"log"
	"os"
)

var GitCommit string
var Version string

func main() {
	log.Printf("s3_terraform_registry. Version: %s; Commit: %s", Version, GitCommit)

	r := gin.Default()

	s3Backend, err := endpoints.NewS3Backend(
		os.Getenv("BUCKET_NAME"),
		os.Getenv("HOSTNAME"),
		os.Getenv("KEYFILE"),
		os.Getenv("KEY_ID"),
	)
	if err != nil {
		panic(err)
	}

	cache := endpoints.NewCache(s3Backend)
	err = cache.Refresh()
	if err != nil {
		panic(err)
	}

	r.GET("/.well-known/terraform.json", endpoints.Discovery())

	r.GET("/v1/providers/:namespace/:type/versions", endpoints.ListVersions(&cache))
	r.GET("/v1/providers/:namespace/:type/:version/download/:os/:arch", endpoints.GetDownloadData(s3Backend))

	r.GET("/proxy/:namespace/:type/:version/:os/:arch/:filename", endpoints.Proxy(s3Backend))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	_ = r.Run(":" + port)
}
