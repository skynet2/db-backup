package configuration

type Configuration struct {
	IncludeDbs    []string                  `env:"INCLUDE_DBS"` // not empty -> include only specified dbs
	ExcludeDbs    []string                  `env:"EXCLUDE_DBS"` // not empty -> exclude databases
	Db            DbConfiguration           `env:"DB"`
	Storage       StorageConfiguration      `env:"STORAGE"`
	Notifications NotificationConfiguration `env:"NOTIFICATIONS"`
	Metrics       Metrics                   `env:"METRICS"`
}

type Metrics struct {
	PrometheusPushGatewayUrl string `yaml:"prometheus_push_gateway_url" env:"PROMETHEUS_PUSH_GATEWAY_URL"`
	PrometheusJobName        string `yaml:"prometheus_job_name" env:"PROMETHEUS_JOB_NAME"`
}

type NotificationConfiguration struct {
	Success NotificationMode `env:"SUCCESS"`
	Fail    NotificationMode `env:"FAIL"`
}

type NotificationMode struct {
	Channels []NotificationChannelConfig `env:"CHANNELS"`
	Template string                      `env:"TEMPLATE"`
}

type DbConfiguration struct {
	Provider string                `env:"PROVIDER"`
	DumpDir  string                `env:"DUMP_DIR"`
	Postgres PostgresConfiguration `env:"POSTGRES"`
}

type StorageConfiguration struct {
	Provider    string   `env:"PROVIDER"`
	DirTemplate string   `env:"DIR_TEMPLATE"`
	MaxFiles    int      `env:"MAX_FILES"`
	S3          S3Config `yaml:"s3" env:"S3"`
}

type PostgresConfiguration struct {
	Host             string `env:"HOST"`
	Port             int    `env:"PORT"`
	User             string `env:"USER"`
	Password         string `env:"PASSWORD"`
	DbDefaultName    string `env:"DB_DEFAULT_NAME"`
	TlsEnabled       bool   `env:"TLS_ENABLED"`
	CompressionLevel int    `env:"COMPRESSION_LEVEL"`
}

type S3Config struct {
	Region         string `env:"REGION"`
	Endpoint       string `env:"ENDPOINT"`
	Bucket         string `env:"BUCKET"`
	AccessKey      string `env:"ACCESS_KEY"`
	SecretKey      string `env:"SECRET_KEY"`
	DisableSsl     bool   `env:"DISABLE_SSL"`
	ForcePathStyle bool   `env:"FORCE_PATH_STYLE"`
}

type NotificationChannelConfig struct {
	Type    string `yaml:"type" env:"NOTIFICATION_CHANNEL_TYPE"`
	Token   string `yaml:"access_token" env:"NOTIFICATION_CHANNEL_ACCESS_TOKEN"`
	Chat    string `yaml:"chat_id" env:"NOTIFICATION_CHANNEL_CHAT_ID"`
	Webhook string `json:"webhook" env:"NOTIFICATION_CHANNEL_WEBHOOK"`
}
