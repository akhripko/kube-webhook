package options

type Config struct {
	LogLevel        string
	HTTPPort        int
	TLSCertFile     string
	TLSKeyFile      string
	HealthCheckPort int
	SystemUsers     []string
	AdminUsers      []string
}
