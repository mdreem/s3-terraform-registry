package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mdreem/s3_terraform_registry/logger"
)

type ListObjects interface {
	ListObjects() ([]string, error)
}

func (bucket Bucket) ListObjects() ([]string, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1"),
	}))

	svc := s3.New(sess)
	objectList, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket.bucketName)})
	if err != nil {
		logger.Sugar.Error("an error occurred when listing versions", "error", err)
		return nil, err
	}

	objects := make([]string, 0)

	for _, item := range objectList.Contents {
		objects = append(objects, *item.Key)
	}

	return objects, err
}
