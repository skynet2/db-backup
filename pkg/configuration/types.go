package configuration

type Configuration struct {
	IncludeDbs    []string // not empty -> include only specified dbs
	ExcludeDbs    []string // not empty -> exclude databases
	Db            DbConfiguration
	Storage       StorageConfiguration
	Notifications NotificationConfiguration
	Metrics       Metrics
}

type Metrics struct {
	PrometheusPushGatewayUrl string `yaml:"prometheus_push_gateway_url" env:"PROMETHEUS_PUSH_GATEWAY_URL"`
	PrometheusJobName        string `yaml:"prometheus_job_name" env:"PROMETHEUS_JOB_NAME"`
}

type NotificationConfiguration struct {
	Success NotificationMode
	Fail    NotificationMode
}

type NotificationMode struct {
	Channels []NotificationChannelConfig
	Template string
}

type DbConfiguration struct {
	Provider string
	DumpDir  string
	Postgres PostgresConfiguration
}

type StorageConfiguration struct {
	Provider    string
	DirTemplate string
	MaxFiles    int
	S3          S3Config `yaml:"s3"`
}

type PostgresConfiguration struct {
	Host             string `env:"POSTGRES_HOST"`
	Port             int    `env:"POSTGRES_PORT"`
	User             string `env:"POSTGRES_USER"`
	Password         string `env:"POSTGRES_PASSWORD"`
	DbDefaultName    string `env:"POSTGRES_DB_DEFAULT_NAME"`
	TlsEnabled       bool   `env:"POSTGRES_TLS_ENABLED"`
	CompressionLevel int    `env:"POSTGRES_COMPRESSION_LEVEL"`
}

type S3Config struct {
	Region         string `env:"S3_REGION"`
	Endpoint       string `env:"S3_ENDPOINT"`
	Bucket         string `env:"S3_BUCKET"`
	AccessKey      string `env:"S3_ACCESS_KEY"`
	SecretKey      string `env:"S3_SECRET_KEY"`
	DisableSsl     bool   `env:"S3_DISABLE_SSL"`
	ForcePathStyle bool   `env:"S3_FORCE_PATH_STYLE"`
}

type NotificationChannelConfig struct {
	Type    string `yaml:"type" env:"NOTIFICATION_CHANNEL_TYPE"`
	Token   string `yaml:"access_token" env:"NOTIFICATION_CHANNEL_ACCESS_TOKEN"`
	Chat    string `yaml:"chat_id" env:"NOTIFICATION_CHANNEL_CHAT_ID"`
	Webhook string `json:"webhook" env:"NOTIFICATION_CHANNEL_WEBHOOK"`
}
