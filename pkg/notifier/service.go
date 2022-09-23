package notifier

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/skynet2/db-backup/pkg/common"
	"github.com/skynet2/db-backup/pkg/configuration"
	"os"
	"strconv"
	"text/template"
)

type DefaultService struct {
	success NotificationInfo
	fail    NotificationInfo
}

func NewDefaultService(cfg configuration.NotificationConfiguration) (Service, error) {
	df := &DefaultService{}

	df.fail.Template = cfg.Fail.Template

	for _, c := range cfg.Fail.Channels {
		if ch, err := createChannel(c); err != nil {
			return nil, errors.WithStack(err)
		} else {
			df.fail.Channels = append(df.fail.Channels, ch)
		}
	}

	df.success.Template = cfg.Success.Template

	for _, c := range cfg.Success.Channels {
		if ch, err := createChannel(c); err != nil {
			return nil, errors.WithStack(err)
		} else {
			df.success.Channels = append(df.success.Channels, ch)
		}
	}

	if len(df.fail.Channels) == 0 && len(df.success.Channels) > 0 {
		df.fail = df.success
	}

	return df, nil
}

func (d *DefaultService) SendResults(ctx context.Context, results []common.Job) error {
	channel := d.success
	enableNotify := false

	if !d.isSuccess(results) {
		channel = d.fail
		enableNotify = true
	}

	msg, err := d.renderTemplate(channel.Template, results)

	if err != nil {
		return err
	}

	var finalErr error

	for _, c := range channel.Channels {
		if err = c.SendMessage(ctx, msg, enableNotify); err != nil {
			finalErr = multierror.Append(finalErr, err)
		}
	}

	return finalErr
}

func (d *DefaultService) SendError(ctx context.Context, err error) error {
	var finalErr error

	for _, ch := range d.fail.Channels {
		if sendErr := ch.SendMessage(ctx, fmt.Sprintf("%+v", err), true); sendErr != nil {
			finalErr = multierror.Append(finalErr, sendErr)
		}
	}

	return finalErr
}

func (d *DefaultService) isSuccess(results []common.Job) bool {
	success := true

	for _, r := range results {
		if r.Error != nil {
			success = false
			break
		}
	}

	return success
}

func (d *DefaultService) renderTemplate(templateStr string, results []common.Job) (string, error) {
	if len(templateStr) == 0 {
		templateStr = `Host {{.host}}
Final result: {{.output}}
Destination: {{.destination}}

Databases:
{{ range $key, $value := .databases }}
{{ $key }}: completed in {{ $value.completed_in}}.{{if $value.size }} Size {{$value.size}}.{{end}} {{ if $value.error }}Error : $value.error {{end}}
{{ end }}`
	}

	compiled, err := template.New("dir").Parse(templateStr)

	if err != nil {
		return "", errors.WithStack(err)
	}

	var buf bytes.Buffer

	hostName, _ := os.Hostname()

	if len(hostName) == 0 {
		hostName = "unk"
	}

	templateParameters := map[string]interface{}{
		"host": hostName,
	}

	dbs := map[string]interface{}{}

	for _, j := range results {
		if len(j.StorageProviderType) > 0 {
			templateParameters["destination"] = j.StorageProviderType
		}

		item := map[string]interface{}{
			"completed_in":        j.EndAt.Sub(j.StartedAt).String(),
			"size":                d.byteCountSI(j.FileSize),
			"backup_completed_in": j.DatabaseBackupEndedAt.Sub(j.DatabaseBackupStartedAt).String(),
		}

		if j.Error != nil {
			item["error"] = fmt.Sprintf("%+v", j.Error)
		}

		dbs[j.DatabaseName] = item
	}

	templateParameters["output"] = strconv.FormatBool(d.isSuccess(results))
	templateParameters["databases"] = dbs

	if err = compiled.Execute(&buf, templateParameters); err != nil {
		return "", errors.WithStack(err)
	}

	return buf.String(), nil
}

func (d *DefaultService) byteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
