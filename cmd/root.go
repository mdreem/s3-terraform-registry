package cmd

import (
	"fmt"
	"github.com/mdreem/s3_terraform_registry/common"
	"github.com/mdreem/s3_terraform_registry/endpoints"
	"github.com/mdreem/s3_terraform_registry/logger"
	"github.com/mdreem/s3_terraform_registry/s3"
	"github.com/spf13/cobra"
	"os"
)

var GitCommit string
var Version string

var RootCmd = &cobra.Command{
	Use: "s3-terraform-registry",
	Run: runCommand,
}

func runCommand(command *cobra.Command, _ []string) {
	if handleVersionFlag(command) {
		return
	}
	logger.Sugar.Infow("s3_terraform_registry. ", "Version", Version, "Commit", GitCommit)

	bucketName := common.GetString(command, "bucket-name")
	hostname := common.GetString(command, "hostname")

	bucket := s3.New(bucketName)
	s3Backend, err := endpoints.NewS3Backend(bucket, hostname)
	if err != nil {
		logger.Sugar.Panicw("failed to initialize S3 backend.", "error", err)
	}

	cache := endpoints.NewCache(s3Backend, bucket)
	err = cache.Refresh()
	if err != nil {
		panic(err)
	}

	r := endpoints.SetupRouter(cache)

	port := common.GetString(command, "port")
	_ = r.Run(":" + port)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Printf("could not execute command: %v", err)
		os.Exit(1)
	}
}

func init() {
	flags := RootCmd.PersistentFlags()
	flags.StringP("bucket-name", "b", "", "the S3 bucket where the files are placed.")

	flags.StringP("hostname", "H", "", "hostname under which this registry will be available.")
	flags.StringP("port", "p", "8080", "port the registry will listen on.")

	flags.StringP("loglevel", "l", "info", "can be set to `error`, `info`, `debug` to set loglevel.")

	markPersistentFlagRequired("bucket-name")
	markPersistentFlagRequired("hostname")
}

func markPersistentFlagRequired(flagName string) {
	err := RootCmd.MarkPersistentFlagRequired(flagName)
	if err != nil {
		logger.Sugar.Errorw("unable to set flag to required.", "flag", flagName)
		os.Exit(1)
	}
}

func handleVersionFlag(c *cobra.Command) bool {
	printVersion := common.GetBoolean(c, "version")
	if printVersion {
		fmt.Printf("\nVersion: %s\n", Version)
		fmt.Printf("Commit:  %s\n", GitCommit)
		return true
	}
	return false
}
