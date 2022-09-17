package storage

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
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
	_, err := s.ListFiles(ctx, s.s3Dir)

	return err
}

func (s S3Provider) ListFiles(ctx context.Context, directory string) ([]string, error) {
	bucket := viper.GetString("S3_BUCKET")

	if len(bucket) == 0 {
		return nil, errors.New("S3_BUCKET is empty")
	}

	ses, err := session.NewSession(s.config)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	cl := s3.New(ses)

	resp, err := cl.ListObjects(&s3.ListObjectsInput{
		Bucket: &bucket,
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}
}
