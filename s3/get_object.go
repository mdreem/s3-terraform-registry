package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mdreem/s3_terraform_registry/logger"
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
	sess := CreateSession(bucket.region)
	svc := s3.New(sess)

	object, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		logger.Sugar.Errorw("an error occurred when getting object", "error", err)
		return BucketObject{}, err
	}

	return BucketObject{
		Body:          object.Body,
		ContentLength: *object.ContentLength,
		ContentType:   *object.ContentType,
	}, err
}
