package main

import
// "flag"
(
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexshnup/wb-shtrixmqtt/conf"
	"github.com/alexshnup/wb-shtrixmqtt/service"
	"github.com/alexshnup/wb-shtrixmqtt/shtrixm"
)

// checkError check error
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	fmt.Printf("%v", conf.Config.Mqtt.Address)

	// interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// open mqtt connection
	client, err := service.NewMqttClient(
		conf.Config.Mqtt.Protocol,
		conf.Config.Mqtt.Address,
		conf.Config.Mqtt.Port,
		0,
	)
	checkError(err)

	// new instance of mqtt client
	wb := shtrixm.NewShtrixm(client, conf.Config.Name, conf.Config.Debug)

	// Run publisher

	// wb.System.Memory.Publish(Config.Timeout, 0)

	// Run subscribing
	wb.System.Shtrixmcmd.Subscribe(2)
	// wb.System.Shtrixmcmd1.Subscribe(2)
	// wb.System.Shtrixmcmd2.Subscribe(2)
	// wb.System.Shtrixmcmd3.Subscribe(2)
	// wb.System.Shtrixmcmd4.Subscribe(2)

	// wait for terminating
	for {
		select {
		case <-interrupt:
			log.Println("Clean and terminating...")

			// Unsubscribe when terminating
			wb.System.Shtrixmcmd.UnSubscribe()
			// wb.System.Shtrixmcmd1.UnSubscribe()
			// wb.System.Shtrixmcmd2.UnSubscribe()
			// wb.System.Shtrixmcmd3.UnSubscribe()
			// wb.System.Shtrixmcmd4.UnSubscribe()

			// disconnecting
			client.Disconnect(250)

			os.Exit(0)
		}
	}

	// LoadJSONConfig()
	//
	// fmt.Println(ConfigJSON.Address)

	// for _, m := range ConfigJSON.Shtrixmcmds {
	// 	fmt.Println(m)
	// 	Shtrixmcmd_ON(ConfigJSON.Address, m)
	// 	StatusShtrixmcmd(ConfigJSON.Address, m)
	// 	time.Sleep(100 * time.Millisecond)
	// 	Shtrixmcmd_OFF(ConfigJSON.Address, m)
	// 	StatusShtrixmcmd(ConfigJSON.Address, m)
	// }

}
