//go:build testing

package testsupport

import (
	"errors"
	"github.com/mdreem/s3_terraform_registry/logger"
	"github.com/mdreem/s3_terraform_registry/s3"
	"github.com/mdreem/s3_terraform_registry/schema"
	"io"
	"strings"
)

type TestBucket struct {
	entries []string
	objects map[string]s3.BucketObject
}

func NewTestBucket(entries []string) TestBucket {
	return TestBucket{entries: entries}
}

func (bucket TestBucket) ListObjects() ([]string, error) {
	return bucket.entries, nil
}

func (bucket TestBucket) GetObject(key string) (s3.BucketObject, error) {
	object, ok := bucket.objects[key]
	if ok {
		return object, nil
	}

	stringReader := strings.NewReader("Object Data for: " + key)
	stringReadCloser := io.NopCloser(stringReader)

	return s3.BucketObject{
		Body:          stringReadCloser,
		ContentLength: int64(len(key)),
		ContentType:   "ContentType: " + key,
	}, nil
}

func NewTestBucketWithObjects(entries []string, objects map[string]s3.BucketObject) TestBucket {
	return TestBucket{entries: entries, objects: objects}
}

func CreateReaderFor(content string) io.ReadCloser {
	stringReader := strings.NewReader(content)
	stringReadCloser := io.NopCloser(stringReader)
	return stringReadCloser
}

func NewTestProviderData() TestProviderData {
	return TestProviderData{}
}

type TestProviderData struct {
}

func (t TestProviderData) ListVersions(namespace string, providerType string) (schema.ProviderVersions, error) {
	logger.Sugar.Infow("listing versions", "namespace", namespace, "type", providerType)
	if namespace == "ERROR_PROVIDER" {
		return schema.ProviderVersions{}, errors.New("some error occurred")
	}

	return schema.ProviderVersions{
		ID: namespace,
		Versions: []schema.ProviderVersion{
			{
				Version:   "1.0.0",
				Protocols: []string{"4.0", "5.0"},
				Platforms: []schema.Platform{{
					Os:   "linux",
					Arch: "amd64",
				}},
			},
		},
		Warnings: nil,
	}, nil
}

func (t TestProviderData) GetDownloadData(namespace string, providerType string, version string, os string, arch string) (schema.DownloadData, error) {
	return schema.DownloadData{}, nil
}

func (t TestProviderData) Proxy(namespace string, providerType string, version string, filename string) (schema.ProxyResponse, error) {
	return schema.ProxyResponse{}, nil
}
