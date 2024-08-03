package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"strconv"

	"github.com/avast/retry-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"github.com/skynet2/db-backup/pkg/configuration"
)

var (
	maxPartSize      = int64(1 * 1024 * 1024 * 1024) // 1 GB
	minMultipartSize = int64(3 * 1024 * 1024 * 1024) // 3 GB
)

const (
	maxRetries = 3
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

func (s S3Provider) Upload(
	ctx context.Context,
	finalFilePath string,
	file *os.File,
) error {
	fileStat, err := file.Stat()
	if err != nil {
		return err
	}

	attempt := 0
	return retry.Do(func() error {
		attempt += 1

		ctx = zerolog.Ctx(ctx).With().Str("file", finalFilePath).
			Str("finalFilePath", finalFilePath).
			Int64("size", fileStat.Size()).
			Int("attempt", attempt).Logger().WithContext(ctx)

		return s.uploadInternal(ctx, finalFilePath, fileStat, file)
	}, retry.Context(ctx), retry.Attempts(5))
}

func (s S3Provider) uploadInternal(
	ctx context.Context,
	finalFilePath string,
	fileStat os.FileInfo,
	file *os.File,
) error {
	if fileStat.Size() < minMultipartSize {
		zerolog.Ctx(ctx).Info().Msgf("Uploading file %v using simple upload", finalFilePath)
		return s.simpleUpload(ctx, finalFilePath, file)
	}

	zerolog.Ctx(ctx).Info().Msgf("Uploading file %s using multipart upload", finalFilePath)
	return s.multiPartUpload(ctx, finalFilePath, file)
}

func (s S3Provider) multiPartUpload(
	ctx context.Context,
	finalFilePath string,
	reader *os.File,
) error {
	cl, err := s.getClient()
	if err != nil {
		return err
	}

	input := &s3.CreateMultipartUploadInput{
		Bucket:      &s.s3Cfg.Bucket,
		Key:         lo.ToPtr(finalFilePath),
		ContentType: lo.ToPtr("application/binary"),
	}

	resp, err := cl.CreateMultipartUpload(input)
	if err != nil {
		return err
	}

	buffer := make([]byte, maxPartSize)
	partNumber := 1
	completedParts := make([]*s3.CompletedPart, 0)
	for {
		n, readErr := reader.Read(buffer)
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}

			return readErr
		}

		toUpload := buffer
		if int64(n) < maxPartSize {
			toUpload = buffer[:n]
		}

		part, uploadPartErr := s.uploadPart(ctx, cl, resp, toUpload, partNumber)
		if uploadPartErr != nil {
			err = s.abortMultipartUpload(ctx, cl, resp)
			if err != nil {
				uploadPartErr = errors.Join(uploadPartErr, errors.WithStack(err))
			}

			return uploadPartErr
		}
		completedParts = append(completedParts, part)
		partNumber += 1
	}

	return s.completeMultipartUpload(ctx, cl, resp, completedParts)
}

func (s S3Provider) completeMultipartUpload(
	ctx context.Context,
	svc *s3.S3,
	resp *s3.CreateMultipartUploadOutput,
	completedParts []*s3.CompletedPart,
) error {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}

	_, err := svc.CompleteMultipartUploadWithContext(ctx, completeInput)
	return err
}

func (s S3Provider) abortMultipartUpload(
	ctx context.Context,
	svc *s3.S3,
	resp *s3.CreateMultipartUploadOutput,
) error {
	zerolog.Ctx(ctx).Warn().Msg("Aborting multipart upload for UploadId#" + *resp.UploadId)
	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   resp.Bucket,
		Key:      resp.Key,
		UploadId: resp.UploadId,
	}

	_, err := svc.AbortMultipartUploadWithContext(ctx, abortInput)
	return err
}

func (s S3Provider) uploadPart(
	ctx context.Context,
	svc *s3.S3,
	resp *s3.CreateMultipartUploadOutput,
	fileBytes []byte,
	partNumber int,
) (*s3.CompletedPart, error) {
	tryNum := 1
	partInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(fileBytes),
		Bucket:        resp.Bucket,
		Key:           resp.Key,
		PartNumber:    aws.Int64(int64(partNumber)),
		UploadId:      resp.UploadId,
		ContentLength: aws.Int64(int64(len(fileBytes))),
	}

	for tryNum <= maxRetries {
		uploadResult, err := svc.UploadPartWithContext(ctx, partInput)
		if err != nil {
			if tryNum == maxRetries {
				var awsErr awserr.Error
				if errors.As(err, &awsErr) {
					return nil, awsErr
				}

				return nil, err
			}
			zerolog.Ctx(ctx).Warn().Msg("Retrying to upload part #" + strconv.Itoa(partNumber))
			tryNum++
		} else {
			zerolog.Ctx(ctx).Debug().Msgf("Uploaded part %v for %v", partNumber, *resp.Key)

			return &s3.CompletedPart{
				ETag:       uploadResult.ETag,
				PartNumber: aws.Int64(int64(partNumber)),
			}, nil
		}
	}

	return nil, nil
}

func (s S3Provider) simpleUpload(
	ctx context.Context,
	finalFilePath string,
	reader io.ReadSeeker,
) error {
	cl, err := s.getClient()
	if err != nil {
		return err
	}

	_, err = cl.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Key:         &finalFilePath,
		Bucket:      &s.s3Cfg.Bucket,
		Body:        reader,
		ContentType: lo.ToPtr("application/binary"),
	})

	return err
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
