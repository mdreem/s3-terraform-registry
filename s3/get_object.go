package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
)

type GetObject interface {
	GetObject(key string) (BucketObject, error)
}

type BucketObject struct {
	Body          io.ReadCloser
	ContentLength int64
	ContentType   string
}

func (bucket Bucket) GetObject(key string) (BucketObject, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1"),
	}))
	svc := s3.New(sess)

	object, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket.bucketName),
		Key:    aws.String(key),
	})
	return BucketObject{
		Body:          object.Body,
		ContentLength: *object.ContentLength,
		ContentType:   *object.ContentType,
	}, err
}
