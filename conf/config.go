// Package conf for go-config-manager
package conf

import (
	"log"

	yamlCfg "github.com/alexshnup/go-config-manager/yaml"
)

// Config struct
type ConfigRoot struct {
	Debug   bool    `yaml:"debug"`
	Timeout int     `yaml:"timeout"`
	Name    string  `yaml:"name"`
	Mqtt    Mqtt    `yaml:"mqtt"`
	Shtrixm Shtrixm `yaml:"shtrixm"`
}

// Mqtt struct
type Mqtt struct {
	Protocol string `yaml:"protocol"`
	Address  string `yaml:"address"`
	Port     string `yaml:"port"`
}

// Shtrixm
type Shtrixm struct {
	Port        int `yaml:"port"`
	ReadTimeout int `yaml:"readtimeout"`
	// ReadTimeout time.Duration `yaml:"readtimeout"`
}

// checkError check error
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Config object
var Config ConfigRoot

func init() {
	// Config manager
	err := yamlCfg.NewConfig("wb-shtrixmqtt-conf.yaml").Load(&Config)
	checkError(err)
}
