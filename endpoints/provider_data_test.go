package endpoints

import (
	"github.com/mdreem/s3_terraform_registry/s3"
	"github.com/mdreem/s3_terraform_registry/schema"
	"io"
	"reflect"
	"strings"
	"testing"
)

func createReaderFor(content string) io.ReadCloser {
	stringReader := strings.NewReader(content)
	stringReadCloser := io.NopCloser(stringReader)
	return stringReadCloser
}

func TestRegistryClient_GetDownloadData(t *testing.T) {
	type fields struct {
		bucket       s3.BucketReaderWriter
		hostname     string
		gpgPublicKey string
		keyID        string
	}
	type args struct {
		namespace    string
		providerType string
		version      string
		os           string
		arch         string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    schema.DownloadData
		wantErr bool
	}{
		{
			name: "get download data",
			fields: fields{
				bucket: NewTestBucketWithObjects([]string{}, map[string]s3.BucketObject{
					"black/lodge/1.0.1/shasum": {
						Body:          createReaderFor("315 coffee"),
						ContentLength: 0,
						ContentType:   "",
					},
				}),
				hostname:     "twin.peaks",
				gpgPublicKey: "Great Northern Hotel Room Key",
				keyID:        "315",
			},
			args: args{
				namespace:    "black",
				providerType: "lodge",
				version:      "1.0.1",
				os:           "linux",
				arch:         "amd64",
			},
			want: schema.DownloadData{
				Protocols:           []string{"4.0", "5.0"},
				Os:                  "linux",
				Arch:                "amd64",
				Filename:            "terraform-provider-lodge_1.0.1_linux_amd64.zip",
				DownloadURL:         "https://twin.peaks/proxy/black/lodge/1.0.1/terraform-provider-lodge_1.0.1_linux_amd64.zip",
				ShasumsURL:          "https://twin.peaks/proxy/black/lodge/1.0.1/shasum",
				ShasumsSignatureURL: "https://twin.peaks/proxy/black/lodge/1.0.1/shasum.sig",
				Shasum:              "315",
				SigningKeys: struct {
					GpgPublicKeys []schema.GpgPublicKey `json:"gpg_public_keys"`
				}{
					GpgPublicKeys: []schema.GpgPublicKey{
						{
							KeyID:          "315",
							ASCIIArmor:     "Great Northern Hotel Room Key",
							TrustSignature: "",
							Source:         "",
							SourceURL:      "",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := RegistryClient{
				bucket:       tt.fields.bucket,
				hostname:     tt.fields.hostname,
				gpgPublicKey: tt.fields.gpgPublicKey,
				keyID:        tt.fields.keyID,
			}
			got, err := client.GetDownloadData(tt.args.namespace, tt.args.providerType, tt.args.version, tt.args.os, tt.args.arch)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDownloadData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDownloadData() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistryClient_ListVersions(t *testing.T) {
	type fields struct {
		bucket       s3.BucketReaderWriter
		hostname     string
		gpgPublicKey string
		keyID        string
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
			name: "list versions based on S3 content",
			fields: fields{
				bucket: NewTestBucket([]string{
					"black/lodge/1.0.0/provider_1.0.0_linux_amd64.zip",
					"black/lodge/1.0.1/provider_1.0.1_linux_amd64.zip",
					"black/lodge/1.0.1/provider_1.0.1_windows_amd64.zip",
				}),
				hostname:     "twin.peaks",
				gpgPublicKey: "Great Northern Hotel Room Key",
				keyID:        "315",
			},
			args: args{
				namespace:    "black",
				providerType: "lodge",
			},
			want: schema.ProviderVersions{
				ID: "black/lodge",
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
						Platforms: []schema.Platform{
							{
								Os:   "linux",
								Arch: "amd64",
							},
							{
								Os:   "windows",
								Arch: "amd64",
							}},
					},
				},
				Warnings: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := RegistryClient{
				bucket:       tt.fields.bucket,
				hostname:     tt.fields.hostname,
				gpgPublicKey: tt.fields.gpgPublicKey,
				keyID:        tt.fields.keyID,
			}
			got, err := client.ListVersions(tt.args.namespace, tt.args.providerType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListVersions() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistryClient_Proxy(t *testing.T) {
	type fields struct {
		bucket       s3.BucketReaderWriter
		hostname     string
		gpgPublicKey string
		keyID        string
	}
	type args struct {
		namespace    string
		providerType string
		version      string
		os           string
		arch         string
		filename     string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ProxyResponse
		wantErr bool
	}{
		{
			name: "proxy returns file",
			fields: fields{
				bucket: NewTestBucketWithObjects([]string{}, map[string]s3.BucketObject{
					"black/lodge/1.0.1/provider_1.0.1_linux_amd64.zip": {
						Body:          createReaderFor("315 coffee"),
						ContentLength: 253,
						ContentType:   "Lodge Response",
					},
				}),
				hostname:     "twin.peaks",
				gpgPublicKey: "Great Northern Hotel Room Key",
				keyID:        "315",
			},
			args: args{
				namespace:    "black",
				providerType: "lodge",
				version:      "1.0.1",
				os:           "linux",
				arch:         "amd64",
				filename:     "provider_1.0.1_linux_amd64.zip",
			},
			want: ProxyResponse{
				Body:          createReaderFor("315 coffee"),
				ContentLength: 253,
				ContentType:   "Lodge Response",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := RegistryClient{
				bucket:       tt.fields.bucket,
				hostname:     tt.fields.hostname,
				gpgPublicKey: tt.fields.gpgPublicKey,
				keyID:        tt.fields.keyID,
			}
			got, err := client.Proxy(tt.args.namespace, tt.args.providerType, tt.args.version, tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("Proxy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Proxy() got = %v, want %v", got, tt.want)
			}
		})
	}
}
