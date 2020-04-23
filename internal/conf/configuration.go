package conf

import (
	"github.com/koding/multiconfig"
)

type Host struct {
	Name     string `required:"true"`
	Address  string `required:"true"`
	Username string `required:"true"`
	Password string
	Key      string
}

// SSHDatasourceDef represents the definition of the ssh datasource
type SSHDatasourceDef struct {
	Host     Host
	JumpHost Host
	File     string
}

// Alias represents a shortcut to a command. It has at least three values: a host, a file and a command.
type Alias struct {
	Datasource interface{} `required:"true"`
	Command    string      `required:"true"`
	Flags      string      `required:"true"`
}

type Configuration struct {
	Hosts       map[string]Host
	Aliases     map[string]Alias
	Datasources map[string]interface{}
}

var (
	configurationFile string
	conf              Configuration
)

func ReadConfiguration(file string) (Configuration, error) {
	l := &multiconfig.TOMLLoader{Path: file}

	c := Configuration{}
	err := l.Load(&c)
	if err != nil {
		return Configuration{}, err
	}
	return c, nil
}
