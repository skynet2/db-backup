package common

import "time"

type Job struct {
	DatabaseName             string
	DatabaseBackupStartedAt  time.Time
	DatabaseBackupEndedAt    time.Time
	StartedAt                time.Time
	EndAt                    time.Time
	StorageFileLocation      string
	StorageProviderStartedAt *time.Time
	UploadEndedAt            *time.Time
	StorageProviderType      string
	Error                    error
	FileLocation             string
	Output                   string
	RemovedFiles             []string
}
