package endpoints

import (
	"errors"
	"fmt"
	"github.com/mdreem/s3_terraform_registry/logger"
	"github.com/mdreem/s3_terraform_registry/s3"
	"github.com/mdreem/s3_terraform_registry/schema"
	"io"
	"reflect"
	"strings"
	"testing"
)

func defaultBucketContent() []string {
	return []string{
		"some_namespace/some_type/1.0.0/",
		"some_namespace/some_type/1.0.0/linux/amd64/terraform-provider-test_1.0.0_linux_amd64.zip",
		"some_namespace/some_type/1.0.1/",
		"some_namespace/some_type/1.0.1/linux/amd64/terraform-provider-test_1.0.1_linux_amd64.zip",
	}
}

func errorBucketContent() []string {
	return []string{
		"ERROR_PROVIDER/some_type/1.0.0/",
		"ERROR_PROVIDER/some_type/1.0.0/linux/amd64/terraform-provider-test_1.0.0_linux_amd64.zip",
	}
}

type TestBucket struct {
	entries []string
	objects map[string]s3.BucketObject
}

func NewTestBucket(entries []string) TestBucket {
	return TestBucket{entries: entries}
}

func NewTestBucketWithObjects(entries []string, objects map[string]s3.BucketObject) TestBucket {
	return TestBucket{entries: entries, objects: objects}
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

func (t TestProviderData) Proxy(namespace string, providerType string, version string, filename string) (ProxyResponse, error) {
	return ProxyResponse{}, nil
}

func defaultVersions() map[string]map[string]schema.ProviderVersions {
	return map[string]map[string]schema.ProviderVersions{
		"replace": {
			"me": {
				ID:       "REPLACEME",
				Versions: nil,
				Warnings: nil,
			},
		},
	}
}

func listVersionDataFor(namespace string, pType string) schema.ProviderVersions {
	return schema.ProviderVersions{
		ID: fmt.Sprintf("%s_%s", namespace, pType),
		Versions: []schema.ProviderVersion{
			{
				Version:   "1.0.0",
				Protocols: []string{"4.0", "5.0"},
				Platforms: []schema.Platform{{
					Os:   "linux",
					Arch: "amd64",
				}},
			},
			{
				Version:   "1.0.1",
				Protocols: []string{"4.0", "5.0"},
				Platforms: []schema.Platform{{
					Os:   "linux",
					Arch: "amd64",
				}},
			},
		},
		Warnings: nil,
	}
}

func listVersionsData() map[string]map[string]schema.ProviderVersions {
	return map[string]map[string]schema.ProviderVersions{
		"black": {
			"lodge": listVersionDataFor("black", "lodge"),
		},
		"white": {
			"lodge": listVersionDataFor("white", "lodge"),
		},
	}
}

func TestS3ProviderData_ListVersions(t *testing.T) {
	type fields struct {
		providerData ProviderData
		cachedResult cachedResult
		bucket       s3.ListObjects
	}
	type args struct {
		namespace    string
		providerType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    schema.ProviderVersions
		wantErr bool
	}{
		{
			name: "fetch cached result for versions data",
			fields: fields{
				providerData: nil,
				cachedResult: cachedResult{
					versions: listVersionsData(),
				},
				bucket: nil,
			},
			args: args{
				namespace:    "black",
				providerType: "lodge",
			},
			want:    listVersionDataFor("black", "lodge"),
			wantErr: false,
		},
		{
			name: "fail to fetch cached result for versions data",
			fields: fields{
				providerData: nil,
				cachedResult: cachedResult{
					versions: listVersionsData(),
				},
				bucket: nil,
			},
			args: args{
				namespace:    "twin",
				providerType: "peaks",
			},
			want:    schema.ProviderVersions{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := S3ProviderData{
				providerData: tt.fields.providerData,
				cachedResult: tt.fields.cachedResult,
				bucket:       tt.fields.bucket,
			}
			got, err := cache.ListVersions(tt.args.namespace, tt.args.providerType)
			if (err != nil) != tt.wantErr {
				t.Errorf("listVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listVersions() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestS3ProviderData_Refresh(t *testing.T) {
	type fields struct {
		providerData ProviderData
		cachedResult cachedResult
		bucket       s3.ListObjects
	}
	tests := []struct {
		name             string
		fields           fields
		wantErr          bool
		wantCachedResult cachedResult
	}{
		{
			name: "test refreshing data in bucket",
			fields: fields{
				providerData: NewTestProviderData(),
				cachedResult: cachedResult{
					versions: defaultVersions(),
				},
				bucket: NewTestBucket(defaultBucketContent()),
			},
			wantErr: false,
			wantCachedResult: cachedResult{
				map[string]map[string]schema.ProviderVersions{
					"some_namespace": {
						"some_type": {
							ID: "some_namespace",
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
						},
					},
				},
			},
		},
		{
			name: "test refreshing data in bucket",
			fields: fields{
				providerData: NewTestProviderData(),
				cachedResult: cachedResult{
					versions: defaultVersions(),
				},
				bucket: NewTestBucket(errorBucketContent()),
			},
			wantErr: true,
			wantCachedResult: cachedResult{
				versions: defaultVersions(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := S3ProviderData{
				providerData: tt.fields.providerData,
				cachedResult: tt.fields.cachedResult,
				bucket:       tt.fields.bucket,
			}
			if err := cache.Refresh(); (err != nil) != tt.wantErr {
				t.Errorf("Refresh() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(cache.cachedResult, tt.wantCachedResult) {
				t.Errorf("Refresh() updated to %v\n, want = %v", cache.cachedResult, tt.wantCachedResult)
			}
		})
	}
}
