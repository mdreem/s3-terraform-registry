package main

import (
	"github.com/mdreem/s3_terraform_registry/endpoints"
	"github.com/mdreem/s3_terraform_registry/logger"
	"github.com/mdreem/s3_terraform_registry/s3"
	"os"
)

var GitCommit string
var Version string

func main() {
	logger.Sugar.Info("s3_terraform_registry. ", "Version", Version, "Commit", GitCommit)

	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		logger.Sugar.Error("BUCKET_NAME not set")
		os.Exit(1)
	}
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		logger.Sugar.Error("HOSTNAME not set")
		os.Exit(1)
	}
	keyfile := os.Getenv("KEYFILE")
	if keyfile == "" {
		logger.Sugar.Error("KEYFILE not set")
		os.Exit(1)
	}
	keyID := os.Getenv("KEY_ID")
	if keyID == "" {
		logger.Sugar.Error("KEY_ID not set")
		os.Exit(1)
	}

	bucket := s3.New(bucketName)
	s3Backend, err := endpoints.NewS3Backend(bucket, hostname, keyfile, keyID)
	if err != nil {
		panic(err)
	}

	cache := endpoints.NewCache(s3Backend, bucket)
	err = cache.Refresh()
	if err != nil {
		panic(err)
	}

	r := endpoints.SetupRouter(cache, s3Backend)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	_ = r.Run(":" + port)
}
