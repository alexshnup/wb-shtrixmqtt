package syscore

import (
	mqtt "github.com/alexshnup/mqtt"
)

// System struct
type System struct {
	Memory     *memory
	Shtrixmcmd *shtrixmcmd
	// Shtrixmcmd1 *shtrixmcmd
	// Shtrixmcmd2 *shtrixmcmd
	// Shtrixmcmd3 *shtrixmcmd
	// Shtrixmcmd4 *shtrixmcmd
}

// NewSystem return new System object.
func NewShtrixM(c mqtt.Client, name string, debug bool) *System {
	return &System{
		// Memory: newMemory(c, name, debug),
		Shtrixmcmd: newShtrixmcmd(c, name, debug),
		// Shtrixmcmd1: newShtrixmcmd(c, name+"/3/1", debug),
		// Shtrixmcmd2: newShtrixmcmd(c, name+"/3/2", debug),
		// Shtrixmcmd3: newShtrixmcmd(c, name+"/3/3", debug),
		// Shtrixmcmd4: newShtrixmcmd(c, name+"/3/4", debug),
	}
}
