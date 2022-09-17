package configuration

type Configuration struct {
	IncludeDbs []string // not empty -> include only specified dbs
	ExcludeDbs []string // not empty -> exclude databases
}
