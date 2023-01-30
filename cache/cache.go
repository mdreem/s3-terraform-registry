package cache

import (
	"fmt"
	"github.com/mdreem/s3_terraform_registry/logger"
	"github.com/mdreem/s3_terraform_registry/providerdata"
	"github.com/mdreem/s3_terraform_registry/s3"
	"github.com/mdreem/s3_terraform_registry/schema"
	"regexp"
)

type Cache interface {
	Refresh() error
}

type CacheableProviderData interface {
	providerdata.ProviderData
	Cache
}

type s3ProviderData struct {
	providerData providerdata.ProviderData
	cachedResult cachedResult
	bucket       s3.ListObjects
}

type cachedResult struct {
	versions map[string]map[string]schema.ProviderVersions
}

func NewCache(client providerdata.ProviderData, bucketReader s3.ListObjects) CacheableProviderData {
	return &s3ProviderData{
		providerData: client,
		cachedResult: cachedResult{},
		bucket:       bucketReader,
	}
}

func (cache s3ProviderData) ListVersions(namespace string, providerType string) (schema.ProviderVersions, error) {
	namespaceData, ok := cache.cachedResult.versions[namespace]
	if !ok {
		return schema.ProviderVersions{}, fmt.Errorf("unable to find data for namespace %s", namespace)
	}

	providerData, ok := namespaceData[providerType]
	if !ok {
		return schema.ProviderVersions{}, fmt.Errorf("unable to find data for type %s", providerType)
	}
	return providerData, nil
}

func (cache s3ProviderData) GetDownloadData(namespace string, providerType string, version string, os string, arch string) (schema.DownloadData, error) {
	return cache.providerData.GetDownloadData(namespace, providerType, version, os, arch)
}

func (cache s3ProviderData) Proxy(namespace string, providerType string, version string, os string) (schema.ProxyResponse, error) {
	return cache.providerData.Proxy(namespace, providerType, version, os)
}

func (cache *s3ProviderData) Refresh() error {
	r := regexp.MustCompile(`^(?P<namespace>[^/]*)/(?P<type>[^/]*)/`)
	names := r.SubexpNames()

	versionData := cachedResult{
		versions: make(map[string]map[string]schema.ProviderVersions),
	}

	objects, err := cache.bucket.ListObjects()
	if err != nil {
		logger.Sugar.Errorw("an error occurred when listing objects in S3", "error", err)
		return err
	}

	for _, item := range objects {
		logger.Sugar.Debugw("checking item", "item", item)
		if r.MatchString(item) {
			result := r.FindAllStringSubmatch(item, -1)
			matches := map[string]string{}
			for i, n := range result[0] {
				matches[names[i]] = n
			}

			listVersions, err := cache.providerData.ListVersions(matches["namespace"], matches["type"])
			if err != nil {
				logger.Sugar.Errorw("an error occurred when updating listing versions", "error", err)
				return err
			}

			_, ok := versionData.versions[matches["namespace"]]
			if !ok {
				versionData.versions[matches["namespace"]] = make(map[string]schema.ProviderVersions)
			}

			versionData.versions[matches["namespace"]][matches["type"]] = listVersions
		}
	}

	cache.cachedResult = versionData
	return nil
}
