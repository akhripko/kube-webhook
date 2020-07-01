package options

import (
	"github.com/spf13/viper"
)

func ReadEnv() *Config {
	viper.AutomaticEnv()

	viper.SetEnvPrefix("APP")

	viper.SetDefault("LOG_LEVEL", "DEBUG")

	viper.SetDefault("HTTP_PORT", 8443)
	viper.SetDefault("INFO_PORT", 8888)

	viper.SetDefault("SYSTEM_USERS", "system:node:docker-desktop system:serviceaccount:kube-system:replicaset-controller")
	viper.SetDefault("ADMIN_USERS", "docker-for-desktop user1 user2")

	return &Config{
		LogLevel:        viper.GetString("LOG_LEVEL"),
		HTTPPort:        viper.GetInt("HTTP_PORT"),
		TLSCertFile:     viper.GetString("TLS_CERT_FILE"),
		TLSKeyFile:      viper.GetString("TLS_KEY_FILE"),
		HealthCheckPort: viper.GetInt("INFO_PORT"),
		SystemUsers:     viper.GetStringSlice("SYSTEM_USERS"),
		AdminUsers:      viper.GetStringSlice("ADMIN_USERS"),
	}
}
