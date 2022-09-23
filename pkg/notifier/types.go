package notifier

import (
	"context"
	"github.com/skynet2/db-backup/pkg/common"
)

type Service interface {
	SendResults(ctx context.Context, results []common.Job) error
	SendError(ctx context.Context, err error) error
}

type NotificationInfo struct {
	Template string
	Channels []Channel
}

type Channel interface {
	SendMessage(ctx context.Context, message string, enableNotification bool) error
}
