package backends

import (
	"bytes"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type s3Persistence struct {
	s3Client *s3.S3
	bucket   string
}

func NewS3(key string, secret string, endpoint string, region string, bucket string) (*s3Persistence, error) {
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(key, secret, ""), // Specifies your credentials.
		Endpoint:    aws.String(endpoint),                              // Find your endpoint in the control panel, under Settings. Prepend "https://".
		Region:      aws.String(region),                                // Must be "us-east-1" when creating new Spaces. Otherwise, use the region in your endpoint, such as "nyc3".
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		return nil, err
	}

	s3Client := s3.New(newSession)

	return &s3Persistence{
		s3Client: s3Client,
		bucket:   bucket,
	}, nil
}

func (p *s3Persistence) Close() {
}

func (p *s3Persistence) UploadFile(fileName string, data []byte) error {
	object := s3.PutObjectInput{
		Bucket: aws.String(p.bucket),      // The path to the directory you want to upload the object to, starting with your Space name.
		Key:    aws.String(fileName),      // Object key, referenced whenever you want to access this file later.
		Body:   bytes.NewReader(data),     // The object's contents.
		ACL:    aws.String("public-read"), // Defines Access-control List (ACL) permissions, such as private or public.
	}

	_, err := p.s3Client.PutObject(&object)
	if err != nil {
		return err
	}

	return nil
}

func (p *s3Persistence) DownloadFile(fileName string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(fileName),
	}

	result, err := p.s3Client.GetObject(input)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
