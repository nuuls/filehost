package s3store

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nuuls/filehost/internal/config"
	"github.com/nuuls/filehost/internal/filestore"
	"github.com/pkg/errors"
)

type S3Store struct {
	client     *s3.Client
	bucketName string
}

var _ filestore.Filestore = &S3Store{}

func New(cfg *config.Config) *S3Store {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: cfg.StorageBucketEndpoint,
		}, nil
	})

	s3cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithEndpointResolverWithOptions(r2Resolver),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.StorageBucketAccessKeyID, cfg.StorageBucketSecretKey, "")),
	)
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.NewFromConfig(s3cfg)

	if err != nil {
		panic(err)
	}
	return &S3Store{
		client:     s3Client,
		bucketName: cfg.StorageBucketName,
	}
}

func (s *S3Store) Get(name string) (io.ReadSeeker, error) {
	obj, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &s.bucketName,
		Key:    &name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "File not found")
	}

	// TODO: fix
	bs, _ := ioutil.ReadAll(obj.Body)
	buf := bytes.NewReader(bs)
	return buf, nil
}

func (s *S3Store) Create(name string, data io.Reader) error {
	// TODO: check if file exists
	// TODO: remove -1 size
	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: &s.bucketName,
		Key:    &name,
		Body:   data,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to create file")
	}

	return nil
}

func (s *S3Store) Delete(name string) error {
	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: &s.bucketName,
		Key:    &name,
	})
	if err != nil {
		errors.Wrap(err, "Failed to delete file")
	}
	return nil
}
