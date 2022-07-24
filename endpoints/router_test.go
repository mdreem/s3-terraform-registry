package endpoints

import (
	"encoding/json"
	"github.com/mdreem/s3_terraform_registry/logger"
	"github.com/mdreem/s3_terraform_registry/s3"
	"github.com/mdreem/s3_terraform_registry/schema"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetVersions(t *testing.T) {
	logger.Logger, _ = zap.NewDevelopment()
	logger.Sugar = logger.Logger.Sugar()

	testBucketWithObjects := NewTestBucketWithObjects([]string{
		"black/lodge/",
		"black/lodge/1.0.0/",
	}, nil)
	providerData := NewTestProviderData()
	cache := NewCache(providerData, testBucketWithObjects)
	err := cache.Refresh()
	if err != nil {
		t.Fatalf("error refreshing cache: %v", err)
	}

	r := SetupRouter(cache, providerData)

	req, _ := http.NewRequest("GET", "/v1/providers/black/lodge/versions", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	providerVersions := schema.ProviderVersions{}
	err = json.Unmarshal(w.Body.Bytes(), &providerVersions)
	if err != nil {
		t.Fatalf("error umarshalling: %v", err)
	}

	wantedProviderVersions := schema.ProviderVersions{
		ID: "black",
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
	}

	if !reflect.DeepEqual(providerVersions, wantedProviderVersions) {
		t.Errorf("fetching versions: got = %v, want %v", providerVersions, wantedProviderVersions)
	}
}

func TestGetDownloadData(t *testing.T) {
	logger.Logger, _ = zap.NewDevelopment()
	logger.Sugar = logger.Logger.Sugar()

	testBucketWithObjects := NewTestBucketWithObjects([]string{
		"black/lodge/",
		"black/lodge/1.0.0/",
		"black/lodge/1.0.1/",
	}, map[string]s3.BucketObject{
		"black/lodge/1.0.1/shasum": {
			Body:          createReaderFor("sha315 coffee"),
			ContentLength: 0,
			ContentType:   "",
		},
		"black/lodge/1.0.1/key_id": {
			Body:          createReaderFor("315"),
			ContentLength: 0,
			ContentType:   "",
		},
		"black/lodge/1.0.1/keyfile": {
			Body:          createReaderFor("Great Northern Hotel Room Key"),
			ContentLength: 0,
			ContentType:   "",
		},
	})
	providerData, err := NewS3Backend(testBucketWithObjects, "twin.peaks")
	if err != nil {
		t.Fatalf("error creating providerData: %v", err)
	}
	cache := NewCache(providerData, testBucketWithObjects)
	err = cache.Refresh()
	if err != nil {
		t.Fatalf("error refreshing cache: %v", err)
	}

	r := SetupRouter(cache, providerData)

	req, _ := http.NewRequest("GET", "/v1/providers/black/lodge/1.0.1/download/linux/amd64", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	downloadData := schema.DownloadData{}
	err = json.Unmarshal(w.Body.Bytes(), &downloadData)
	if err != nil {
		t.Errorf("error umarshalling: %v", err)
	}

	wantedDownloadData := schema.DownloadData{
		Protocols:           []string{"4.0", "5.0"},
		Os:                  "linux",
		Arch:                "amd64",
		Filename:            "terraform-provider-lodge_1.0.1_linux_amd64.zip",
		DownloadURL:         "https://twin.peaks/proxy/black/lodge/1.0.1/terraform-provider-lodge_1.0.1_linux_amd64.zip",
		ShasumsURL:          "https://twin.peaks/proxy/black/lodge/1.0.1/shasum",
		ShasumsSignatureURL: "https://twin.peaks/proxy/black/lodge/1.0.1/shasum.sig",
		Shasum:              "sha315",
		SigningKeys: struct {
			GpgPublicKeys []schema.GpgPublicKey `json:"gpg_public_keys"`
		}{
			GpgPublicKeys: []schema.GpgPublicKey{
				{
					KeyID:      "315",
					ASCIIArmor: "Great Northern Hotel Room Key",
				},
			},
		},
	}

	if !reflect.DeepEqual(downloadData, wantedDownloadData) {
		t.Errorf("fetching download data: got = %v, want %v", downloadData, wantedDownloadData)
	}
}

func TestProxy(t *testing.T) {
	logger.Logger, _ = zap.NewDevelopment()
	logger.Sugar = logger.Logger.Sugar()

	testBucketWithObjects := NewTestBucketWithObjects([]string{
		"black/lodge/",
		"black/lodge/1.0.0/",
		"black/lodge/1.0.1/",
	}, map[string]s3.BucketObject{
		"black/lodge/1.0.1/terraform-provider-lodge_1.0.1_linux_amd64.zip": {
			Body:          createReaderFor("315 coffee provider"),
			ContentLength: 0,
			ContentType:   "",
		},
	})
	providerData, err := NewS3Backend(testBucketWithObjects, "twin.peaks")
	if err != nil {
		t.Fatalf("error creating providerData: %v", err)
	}
	cache := NewCache(providerData, testBucketWithObjects)
	err = cache.Refresh()
	if err != nil {
		t.Fatalf("error refreshing cache: %v", err)
	}

	r := SetupRouter(cache, providerData)

	req, _ := http.NewRequest("GET", "/proxy/black/lodge/1.0.1/terraform-provider-lodge_1.0.1_linux_amd64.zip", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	responseData := w.Body.String()
	const wantedResponse = "315 coffee provider"
	if responseData != wantedResponse {
		t.Errorf("fetching file: got = %v, want %v", responseData, wantedResponse)
	}
}

func TestRefresh(t *testing.T) {
	logger.Logger, _ = zap.NewDevelopment()
	logger.Sugar = logger.Logger.Sugar()

	testBucketWithObjects := NewTestBucketWithObjects([]string{
		"black/lodge/",
		"black/lodge/1.0.0/",
		"black/lodge/1.0.1/",
	}, map[string]s3.BucketObject{
		"black/lodge/1.0.1/terraform-provider-lodge_1.0.1_linux_amd64.zip": {
			Body:          createReaderFor("315 coffee provider"),
			ContentLength: 0,
			ContentType:   "",
		},
	})
	providerData, err := NewS3Backend(testBucketWithObjects, "twin.peaks")
	if err != nil {
		t.Fatalf("error creating providerData: %v", err)
	}
	cache := NewCache(providerData, testBucketWithObjects)

	_, err = cache.ListVersions("black", "lodge")
	if err == nil {
		t.Errorf("expected error here as cache is empty")
	}

	r := SetupRouter(cache, providerData)
	req, _ := http.NewRequest("GET", "/refresh", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	responseData := w.Body.String()
	const wantedResponse = "refreshed"
	if responseData != wantedResponse {
		t.Errorf("refreshing cache: got = %v, want %v", responseData, wantedResponse)
	}

	versions, err := cache.ListVersions("black", "lodge")
	if err != nil {
		return
	}

	const wantedVersionsID = "black/lodge"
	if versions.ID != wantedVersionsID {
		t.Errorf("fetching cached data: got = %v, want %v", versions.ID, wantedVersionsID)
	}
}
