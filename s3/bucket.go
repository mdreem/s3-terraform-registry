package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

type BucketReaderWriter interface {
	ListObjects
	GetObject
}

type Bucket struct {
	region     string
	bucketName string
}

func New(region string, bucketName string) Bucket {
	return Bucket{region: region, bucketName: bucketName}
}

var CreateSession = func(region string) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
}
