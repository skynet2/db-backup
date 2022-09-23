package storage

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/skynet2/db-backup/pkg/configuration"
	"io"
)

type S3Provider struct {
	config *aws.Config
	s3Cfg  configuration.S3Config
}

func NewS3Provider(cfg configuration.S3Config) Provider {
	config := &aws.Config{}

	if region := cfg.Region; len(region) > 0 {
		config.Region = aws.String(region)
	}

	if endpoint := cfg.Endpoint; len(endpoint) > 0 {
		config.Endpoint = aws.String(endpoint)
	}

	if disableSSL := cfg.DisableSsl; disableSSL {
		config.DisableSSL = aws.Bool(disableSSL)
	}

	if s3ForcePath := cfg.ForcePathStyle; s3ForcePath {
		config.S3ForcePathStyle = aws.Bool(s3ForcePath)
	}

	accessKey := cfg.AccessKey
	secretKey := cfg.SecretKey

	if len(accessKey) > 0 || len(secretKey) > 0 {
		config.Credentials = credentials.NewStaticCredentials(accessKey, secretKey, "")
	}

	final := &S3Provider{
		config: config,
		s3Cfg:  cfg,
	}

	return final
}

func (s S3Provider) Validate(ctx context.Context) error {
	_, err := s.List(ctx, "./")

	return err
}

func (s S3Provider) GetType() string {
	return "s3"
}

func (s S3Provider) Remove(ctx context.Context, absolutePath string) error {
	cl, err := s.getClient()

	if err != nil {
		return errors.WithStack(err)
	}

	_, err = cl.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: &s.s3Cfg.Bucket,
		Key:    &absolutePath,
	})

	return err
}

func (s S3Provider) Upload(ctx context.Context, finalFilePath string, reader io.ReadSeeker) error {
	cl, err := s.getClient()

	if err != nil {
		return errors.WithStack(err)
	}

	cType := "application/binary"

	_, err = cl.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Key:         &finalFilePath,
		Bucket:      &s.s3Cfg.Bucket,
		Body:        reader,
		ContentType: &cType,
	})

	return errors.WithStack(err)
}

func (s S3Provider) getClient() (*s3.S3, error) {
	if len(s.s3Cfg.Bucket) == 0 {
		return nil, errors.New("S3_BUCKET is empty")
	}

	ses, err := session.NewSession(s.config)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return s3.New(ses), nil
}

func (s S3Provider) List(ctx context.Context, prefix string) ([]File, error) {
	cl, err := s.getClient()

	if err != nil {
		return nil, errors.WithStack(err)
	}

	resp, err := cl.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket: &s.s3Cfg.Bucket,
		Prefix: &prefix,
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}

	var finalFiles []File

	for _, c := range resp.Contents {
		if c.Key == nil || c.LastModified == nil {
			continue
		}

		finalFiles = append(finalFiles, File{
			AbsolutePath: *c.Key,
			CreatedAt:    *c.LastModified,
		})
	}

	return sortFiles(finalFiles), nil
}
