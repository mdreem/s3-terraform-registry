//go:build testing

package testsupport

import (
	"github.com/mdreem/s3_terraform_registry/s3"
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
