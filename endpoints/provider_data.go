package endpoints

import (
	"bytes"
	"fmt"
	"github.com/mdreem/s3_terraform_registry/logger"
	"github.com/mdreem/s3_terraform_registry/s3"
	"github.com/mdreem/s3_terraform_registry/schema"
	"io"
	"io/ioutil"
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
	Proxy(namespace string, providerType string, version string, os string, arch string, filename string) (ProxyResponse, error)
}

type RegistryClient struct {
	bucket       s3.BucketReaderWriter
	hostname     string
	gpgPublicKey string
	keyID        string
}

func NewS3Backend(bucketName string, hostname string, keyFile string, keyID string) (RegistryClient, error) {
	file, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return RegistryClient{}, err
	}

	return RegistryClient{
		bucket:       s3.New(bucketName),
		hostname:     hostname,
		gpgPublicKey: string(file),
		keyID:        keyID,
	}, nil
}

func (client RegistryClient) ListVersions(namespace string, providerType string) (schema.ProviderVersions, error) {
	objects, err := client.bucket.ListObjects()
	if err != nil {
		logger.Error("an error occurred when listing objects in S3", "error", err)
		return schema.ProviderVersions{}, err
	}

	r := regexp.MustCompile(`^(?P<namespace>[^/]*)/(?P<type>[^/]*)/(?P<version>[^/]*)/(?P<os>[^/]*)/(?P<arch>[^/]*)/(?P<name>[^/]*).zip$`)
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

			logger.Info("list versions: adding", "item", item)

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
	basePath := fmt.Sprintf("%s/%s/%s/%s/%s", namespace, providerType, version, os, arch)
	baseURL := fmt.Sprintf("https://%s/proxy/%s", client.hostname, basePath)
	logger.Info("fetching signature file", "file", fmt.Sprintf("%s/shasum", basePath))

	object, err := client.bucket.GetObject(fmt.Sprintf("%s/shasum", basePath))
	if err != nil {
		return schema.DownloadData{}, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(object.Body)
	if err != nil {
		return schema.DownloadData{}, err
	}

	shaSumFile := buf.String()
	shaSum := strings.Split(shaSumFile, " ")[0]

	return schema.DownloadData{
		Protocols:           []string{"4.0", "5.0"},
		Os:                  os,
		Arch:                arch,
		Filename:            fmt.Sprintf("terraform-provider-%s.zip", providerType),
		DownloadURL:         fmt.Sprintf("%s/terraform-provider-%s.zip", baseURL, providerType),
		ShasumsURL:          fmt.Sprintf("%s/shasum", baseURL),
		ShasumsSignatureURL: fmt.Sprintf("%s/shasum.sig", baseURL),
		Shasum:              shaSum,
		SigningKeys: struct {
			GpgPublicKeys []schema.GpgPublicKey `json:"gpg_public_keys"`
		}{
			[]schema.GpgPublicKey{
				{
					KeyID:      client.keyID,
					ASCIIArmor: client.gpgPublicKey,
				},
			},
		},
	}, nil
}

func (client RegistryClient) Proxy(namespace string, providerType string, version string, os string, arch string, filename string) (ProxyResponse, error) {
	basePath := fmt.Sprintf("%s/%s/%s/%s/%s", namespace, providerType, version, os, arch)
	logger.Info("proxying file file", "file", fmt.Sprintf("%s/%s", basePath, filename))

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
