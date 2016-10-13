package shtrixm

import (
	mqtt "github.com/alexshnup/mqtt"

	"github.com/alexshnup/wb-shtrixmqtt/shtrixm/core"
)

// Shtrixm struct
type Shtrixm struct {
	client mqtt.Client
	name   string
	System *syscore.System
}

func NewShtrixm(c mqtt.Client, name string, debug bool) *Shtrixm {
	return &Shtrixm{
		client: c,
		name:   name,
		System: syscore.NewShtrixM(c, name, debug),
	}
}
