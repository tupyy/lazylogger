package conf

import (
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Host struct {
	Address  string
	Username string
	Password string
	Key      string
}

func (h *Host) String() string {
	return fmt.Sprintf("%s:%d", h.Address, 22)
}

type ConfigurationEntry struct {
	Name     string `mapstructure:"name"`
	Host     Host   `mapstructure:"host"`
	JumpHost Host   `mapstructure:"jumpHost"`
	File     string
}

type Configuration struct {
	entries []ConfigurationEntry `mapstructure:"entries"`
}

var (
	configurationFile string
	conf              Configuration
)

func ReadConfigurationFile(file string) Configuration {
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		os.Stderr.WriteString(fmt.Sprintf("Configuration error: %s", err))
	}

	err := mapstructure.Decode(viper.AllSettings(), &conf)
	if err != nil {
		panic(err)
	}

	return conf
}
