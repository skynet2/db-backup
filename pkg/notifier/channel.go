package notifier

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/skynet2/db-backup/pkg/configuration"
	"strings"
)

func createChannel(cfg configuration.NotificationChannelConfig) (Channel, error) {
	chType := strings.TrimSpace(strings.ToLower(cfg.Type))

	switch chType {
	case "telegram":
		return NewTelegramChannel(cfg.Chat, cfg.Token), nil
	default:
		return nil, errors.New(fmt.Sprintf("no implementation for notification provider %v", chType))
	}
}
