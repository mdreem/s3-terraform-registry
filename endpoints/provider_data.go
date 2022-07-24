package endpoints

import (
	"bytes"
	"fmt"
	"github.com/mdreem/s3_terraform_registry/logger"
	"github.com/mdreem/s3_terraform_registry/s3"
	"github.com/mdreem/s3_terraform_registry/schema"
	"io"
	"regexp"
	"sort"
	"strings"
)

type ProxyResponse struct {
	Body          io.ReadCloser
	ContentLength int64
	ContentType   string
}

type ProviderData interface {
	ListVersions(namespace string, providerType string) (schema.ProviderVersions, error)
	GetDownloadData(namespace string, providerType string, version string, os string, arch string) (schema.DownloadData, error)
	Proxy(namespace string, providerType string, version string, os string) (ProxyResponse, error)
}

type RegistryClient struct {
	bucket   s3.BucketReaderWriter
	hostname string
}

func NewS3Backend(bucket s3.BucketReaderWriter, hostname string) (RegistryClient, error) {
	return RegistryClient{
		bucket:   bucket,
		hostname: hostname,
	}, nil
}

func (client RegistryClient) ListVersions(namespace string, providerType string) (schema.ProviderVersions, error) {
	objects, err := client.bucket.ListObjects()
	if err != nil {
		logger.Sugar.Errorw("an error occurred when listing objects in S3", "error", err)
		return schema.ProviderVersions{}, err
	}

	r := regexp.MustCompile(`^(?P<namespace>[^/]*)/(?P<type>[^/]*)/(?P<version>[^/]*)/(?P<name>[^_]*)_(?P<version_file>[^_]*)_(?P<os>[^_]*)_(?P<arch>[^_]*).zip$`)
	names := r.SubexpNames()

	versions := make(map[string][]schema.Platform)

	for _, item := range objects {
		if r.MatchString(item) {
			result := r.FindAllStringSubmatch(item, -1)
			matches := map[string]string{}
			for i, n := range result[0] {
				matches[names[i]] = n
			}

			if matches["name"] == "" {
				continue
			}
			if matches["namespace"] != namespace {
				continue
			}
			if matches["type"] != providerType {
				continue
			}

			logger.Sugar.Infow("list versions: adding", "item", item)

			platforms := versions[matches["version"]]
			platforms = append(platforms, schema.Platform{
				Os:   matches["os"],
				Arch: matches["arch"],
			})
			versions[matches["version"]] = platforms
		}
	}

	providerVersions := make([]schema.ProviderVersion, 0)
	for version, versionData := range versions {
		providerVersion := schema.ProviderVersion{
			Version:   version,
			Protocols: []string{"4.0", "5.0"},
			Platforms: versionData,
		}
		providerVersions = append(providerVersions, providerVersion)
	}

	sort.Slice(providerVersions, func(i, j int) bool {
		return providerVersions[i].Version < providerVersions[j].Version
	})

	return schema.ProviderVersions{
		ID:       fmt.Sprintf("%s/%s", namespace, providerType),
		Versions: providerVersions,
		Warnings: nil,
	}, nil
}

func (client RegistryClient) GetDownloadData(namespace string, providerType string, version string, os string, arch string) (schema.DownloadData, error) {
	basePath := fmt.Sprintf("%s/%s/%s", namespace, providerType, version)
	baseURL := fmt.Sprintf("https://%s/proxy/%s", client.hostname, basePath)

	logger.Sugar.Debugw("getting download data with", "basePath", basePath, "baseURL", baseURL)

	shaSum, err := client.fetchShaSum(basePath)
	if err != nil {
		return schema.DownloadData{}, err
	}

	keyIDFileLocation := fmt.Sprintf("%s/key_id", basePath)
	keyID, err := client.fetchObjectAsString(keyIDFileLocation)
	if err != nil {
		return schema.DownloadData{}, err
	}

	keyfileLocation := fmt.Sprintf("%s/keyfile", basePath)
	gpgPublicKey, err := client.fetchObjectAsString(keyfileLocation)
	if err != nil {
		return schema.DownloadData{}, err
	}

	filename := fmt.Sprintf("terraform-provider-%s_%s_%s_%s.zip", providerType, version, os, arch)
	return schema.DownloadData{
		Protocols:           []string{"4.0", "5.0"},
		Os:                  os,
		Arch:                arch,
		Filename:            filename,
		DownloadURL:         fmt.Sprintf("%s/%s", baseURL, filename),
		ShasumsURL:          fmt.Sprintf("%s/shasum", baseURL),
		ShasumsSignatureURL: fmt.Sprintf("%s/shasum.sig", baseURL),
		Shasum:              shaSum,
		SigningKeys: struct {
			GpgPublicKeys []schema.GpgPublicKey `json:"gpg_public_keys"`
		}{
			[]schema.GpgPublicKey{
				{
					KeyID:      keyID,
					ASCIIArmor: gpgPublicKey,
				},
			},
		},
	}, nil
}

func (client RegistryClient) fetchShaSum(basePath string) (string, error) {
	shaSumLocation := fmt.Sprintf("%s/shasum", basePath)
	logger.Sugar.Debugw("fetching signature file", "file", shaSumLocation)

	shaSumFile, err := client.fetchObjectAsString(shaSumLocation)
	if err != nil {
		return "", err
	}
	shaSum := strings.Split(shaSumFile, " ")[0]
	return shaSum, nil
}

func (client RegistryClient) fetchObjectAsString(objectLocation string) (string, error) {
	object, err := client.bucket.GetObject(objectLocation)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(object.Body)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (client RegistryClient) Proxy(namespace string, providerType string, version string, filename string) (ProxyResponse, error) {
	basePath := fmt.Sprintf("%s/%s/%s", namespace, providerType, version)
	logger.Sugar.Infow("proxying file file", "file", fmt.Sprintf("%s/%s", basePath, filename))

	object, err := client.bucket.GetObject(fmt.Sprintf("%s/%s", basePath, filename))
	if err != nil {
		return ProxyResponse{}, err
	}

	return ProxyResponse{
		Body:          object.Body,
		ContentLength: object.ContentLength,
		ContentType:   object.ContentType,
	}, nil
}
