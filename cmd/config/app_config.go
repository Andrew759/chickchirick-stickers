package config

import (
	"chickchirick-stickers/cmd/config/dto"
	globalConfig "chickchirick-stickers/pkg/chirick_config"

	"github.com/spf13/viper"
)

type AppConfigurationInterface interface {
	NewAppConfiguration() AppConfiguration
}

type AppConfiguration struct {
	ServerURL   string
	Environment string
	dto.DatabaseConfig
}

func (c AppConfiguration) NewAppConfiguration() AppConfiguration {
	return AppConfiguration{
		ServerURL:      viper.GetString(globalConfig.ServerUrl),
		Environment:    viper.GetString(globalConfig.Enviroment),
		DatabaseConfig: PrepareDatabaseConfig(),
	}
}

func PrepareDatabaseConfig() dto.DatabaseConfig {
	dbc := dto.DatabaseConfig{}

	dbc.SetHost(viper.GetString(globalConfig.DbHost))
	dbc.SetPort(viper.GetInt(globalConfig.DbInternalPort))
	dbc.SetName(viper.GetString(globalConfig.DbName))
	dbc.SetUser(viper.GetString(globalConfig.DbUser))
	dbc.SetPassword(viper.GetString(globalConfig.DbPass))
	dbc.SetTimezone(viper.GetString(globalConfig.DbTimezone))

	return dbc
}
