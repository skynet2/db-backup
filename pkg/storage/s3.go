package storage

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io"
)

type S3Provider struct {
	config *aws.Config
	s3Dir  string
}

func NewS3Provider() Provider {
	config := &aws.Config{}

	if region := viper.GetString("S3_REGION"); len(region) > 0 {
		config.Region = aws.String(region)
	}

	if endpoint := viper.GetString("S3_ENDPOINT"); len(endpoint) > 0 {
		config.Endpoint = aws.String(endpoint)
	}

	if disableSSL := viper.GetBool("S3_DISABLE_SSL"); disableSSL {
		config.DisableSSL = aws.Bool(disableSSL)
	}

	if s3ForcePath := viper.GetBool("S3_FORCE_PATH_STYLE"); s3ForcePath {
		config.S3ForcePathStyle = aws.Bool(s3ForcePath)
	}

	accessKey := viper.GetString("S3_ACCESS_KEY")
	secretKey := viper.GetString("S3_SECRET_KEY")

	if len(accessKey) > 0 || len(secretKey) > 0 {
		config.Credentials = credentials.NewStaticCredentials(accessKey, secretKey, "")
	}

	final := &S3Provider{
		config: config,
		s3Dir:  viper.GetString("S3_DIR"),
	}

	return final
}

func (s S3Provider) Validate(ctx context.Context) error {
	_, err := s.List(ctx, s.s3Dir)

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

	bucket := s.getBucket()

	_, err = cl.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &absolutePath,
	})

	return err
}

func (s S3Provider) Upload(ctx context.Context, finalFilePath string, reader io.Reader) error {
	//TODO implement me
	panic("implement me")
}

func (s S3Provider) getClient() (*s3.S3, error) {
	bucket := s.getBucket()

	if len(bucket) == 0 {
		return nil, errors.New("S3_BUCKET is empty")
	}

	ses, err := session.NewSession(s.config)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return s3.New(ses), nil
}

func (s S3Provider) getBucket() string {
	return viper.GetString("S3_BUCKET")
}

func (s S3Provider) List(ctx context.Context, prefix string) ([]File, error) {
	cl, err := s.getClient()

	if err != nil {
		return nil, errors.WithStack(err)
	}

	bucket := s.getBucket()

	resp, err := cl.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket: &bucket,
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
