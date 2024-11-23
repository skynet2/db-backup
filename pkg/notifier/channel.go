package notifier

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/skynet2/db-backup/pkg/configuration"
)

func createChannel(cfg configuration.NotificationChannelConfig) (Channel, error) {
	chType := strings.TrimSpace(strings.ToLower(cfg.Type))

	switch chType {
	case "telegram":
		return NewTelegramChannel(cfg.Chat, cfg.Token), nil
	case "discord":
		return NewDiscordChannel(cfg.Webhook), nil
	case "mattermost":
		return NewMattermost(cfg.MattermostWebhook), nil
	default:
		return nil, errors.New(fmt.Sprintf("no implementation for notification provider %v", chType))
	}
}
