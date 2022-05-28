package s3

type BucketReaderWriter interface {
	ListObjects
	GetObject
}

type Bucket struct {
	bucketName string
}

func New(bucketName string) Bucket {
	return Bucket{bucketName: bucketName}
}
