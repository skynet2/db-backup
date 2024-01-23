package storage

//func TestS3MultipartUpload(t *testing.T) {
//	srv := NewS3Provider(configuration.S3Config{
//
//		DisableSsl:     false,
//		ForcePathStyle: false,
//	})
//
//	file, err := os.Open("/home/iqpirat/Downloads/docker-desktop-4.26.1-amd64.deb")
//	assert.NoError(t, err)
//
//	maxPartSize = 100 * 1024 * 1024      // 100 MB
//	minMultipartSize = 300 * 1024 * 1024 // 300 MB
//
//	zerolog.DefaultContextLogger = lo.ToPtr(zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger())
//	err = srv.Upload(context.TODO(), "docker-desktop-4.26.1-amd64.deb", file)
//	assert.NoError(t, err)
//}
