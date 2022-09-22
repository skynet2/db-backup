package configuration

type Configuration struct {
	IncludeDbs []string // not empty -> include only specified dbs
	ExcludeDbs []string // not empty -> exclude databases
	Db         DbConfiguration
	Storage    StorageConfiguration
}

type DbConfiguration struct {
	Provider string
	DumpDir  string
	Postgres PostgresConfiguration
}

type StorageConfiguration struct {
	Provider string
	Dir      string
	MaxFiles int
	S3       S3Config `yaml:"s3"`
}

type PostgresConfiguration struct {
	Host             string
	Port             int
	User             string
	Password         string
	DbDefaultName    string
	TlsEnabled       bool
	CompressionLevel int
}

type S3Config struct {
	Region    string
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string

	DisableSsl     bool
	ForcePathStyle bool
}
