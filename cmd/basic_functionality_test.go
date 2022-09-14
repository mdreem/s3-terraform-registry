package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/docker/go-connections/nat"
	registryS3 "github.com/mdreem/s3_terraform_registry/s3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const bucketName = "testbucket"
const region = "eu-east-1"

type localstackContainer struct {
	testcontainers.Container
	Endpoint string
}

func setupLocalstack(ctx context.Context) (*localstackContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "localstack/localstack:1.0.2",
		ExposedPorts: []string{"4566/tcp", "4510-4559/tcp"},
		WaitingFor:   wait.ForListeningPort("4566/tcp"),
		Env: map[string]string{
			"SERVICES": "s3",
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	hostname, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	containerPort, err := nat.NewPort("tcp", "4566")
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, containerPort)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("http://%v:%v", hostname, port.Port())

	return &localstackContainer{
		Container: container,
		Endpoint:  endpoint,
	}, nil
}

func getSessionCreator(localstack localstackContainer) func(region string) *session.Session {
	return func(region string) *session.Session {
		sess, _ := session.NewSession(&aws.Config{
			Region:           aws.String(region),
			Credentials:      credentials.NewStaticCredentials("test", "test", ""),
			S3ForcePathStyle: aws.Bool(true),
			Endpoint:         aws.String(localstack.Endpoint),
		})
		return sess
	}
}

func TestIntegrationBasicFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	localstack, err := setupLocalstack(ctx)
	if err != nil {
		t.Fatalf("failed to start container: %v\n", err)
	}
	defer func() {
		if err := localstack.Terminate(ctx); err != nil {
			t.Logf("error terminating localstack: %v\n", err)
		}
	}()

	t.Logf("localstack: %v\n", *localstack)

	sess := getSessionCreator(*localstack)(region)
	registryS3.CreateSession = getSessionCreator(*localstack)

	client := s3.New(sess, aws.NewConfig().WithS3ForcePathStyle(true))

	err = createBucket(client, bucketName)
	if err != nil {
		t.Fatalf("failed to create bucket: %v\n", err)
	}

	createProvider(t, sess, "black", "lodge", "1.0.0", "linux", "amd64", "terraform-provider-lodge")
	createProvider(t, sess, "black", "lodge", "1.0.1", "linux", "amd64", "terraform-provider-lodge")
	createProvider(t, sess, "white", "lodge", "2.0.0", "linux", "amd64", "terraform-provider-lodge")
	createProvider(t, sess, "white", "lodge", "2.0.1", "linux", "amd64", "terraform-provider-lodge")

	RootCmd.SetArgs([]string{
		"--bucket-name", bucketName,
		"--hostname", "twin.peaks.provider",
		"--region", region,
	})

	go runRegistry(t)
	waitToStart(t)

	checkVersionsEndpoint(t)
	checkDownloadsEndpoint(t)
}

func checkVersionsEndpoint(t *testing.T) {
	result, err := getURL(t, "http://localhost:8080/v1/providers/black/lodge/versions")
	if err != nil {
		t.Fatalf("unable to fetch URL: %v\n", err)
	}

	providerID, ok := result["id"]
	if !ok {
		t.Errorf("unable to find key 'id' in response")
	}

	if providerID != "black/lodge" {
		t.Errorf("providerID is '%s' and not 'black/lodge'", providerID)
	}
}

func checkDownloadsEndpoint(t *testing.T) {
	result, err := getURL(t, "http://localhost:8080/v1/providers/white/lodge/2.0.1/download/linux/amd64")
	if err != nil {
		t.Fatalf("unable to fetch URL: %v\n", err)
	}

	t.Logf("result: %v", result)

	downloadURL, ok := result["download_url"]
	if !ok {
		t.Errorf("unable to find key 'download_url' in response")
	}

	if downloadURL != "https://twin.peaks.provider/proxy/white/lodge/2.0.1/terraform-provider-lodge_2.0.1_linux_amd64.zip" {
		t.Errorf("providerID is '%s' and not 'https://twin.peaks.provider/proxy/white/lodge/2.0.1/terraform-provider-lodge_2.0.1_linux_amd64.zip'", downloadURL)
	}
}

func getURL(t *testing.T, url string) (map[string]interface{}, error) {
	resp, err := http.Get(url)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("error closing response body: %v", err)
		}
	}()
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func waitToStart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var body *http.Response
	defer func() {
		if err := body.Body.Close(); err != nil {
			t.Fatalf("unable to close response body: %v\n", err)
		}
	}()
	var err error

	for {
		body, err = http.Get("http://localhost:8080/.well-known/terraform.json")
		if err == nil || ctx.Err() != nil {
			break
		}
	}
}

func runRegistry(t *testing.T) {
	if err := RootCmd.Execute(); err != nil {
		t.Logf("failed to start registry: %v\n", err)
	}
}

//nolint: unparam, gofmt
func createProvider(t *testing.T, sess *session.Session, namespace, providerType, version, os, arch, name string) {
	path := fmt.Sprintf("%s/%s/%s/", namespace, providerType, version)
	zipFileName := fmt.Sprintf("%s_%s_%s_%s.zip", name, version, os, arch)

	uploadFileToS3(t, sess, path+zipFileName)
	uploadFileToS3(t, sess, path+"shasum")
	uploadFileToS3(t, sess, path+"shasum.sig")
	uploadFileToS3(t, sess, path+"key_id")
	uploadFileToS3(t, sess, path+"keyfile")
}

func uploadFileToS3(t *testing.T, sess *session.Session, filename string) {
	fileData := "Content for: " + filename
	bucketFileData := strings.NewReader(fileData)

	uploader := s3manager.NewUploader(sess)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
		Body:   bucketFileData,
	})
	if err != nil {
		t.Fatalf("failed to upload file to bucket: %v\n", err)
	}

	t.Logf("file uploaded to, %s\n", result.Location)
}

func createBucket(client *s3.S3, bucketName string) error {
	_, err := client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})

	return err
}
