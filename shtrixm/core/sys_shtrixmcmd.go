package syscore

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	mqtt "github.com/alexshnup/mqtt"
)

/*
Struct shtrixmcmd provides system LED[0,1] control

Topics:
	Subscribe:
		name + "/SYSTEM/LED[0,1]/ACTION		{0, 1, STATUS}
	Publish:
		name + "/SYSTEM/LED[0,1]/STATUS		{0, 1}

Methods:
	Subscribe
	Unsubscribe
	PublishStatus

Functions:
	Set trigger to [none] when subscribe
		echo none | sudo tee /sys/class/shtrixmcmds/shtrixmcmd0/trigger
	Set trigger to [mmc0] when unsubscribe
		echo mmc0 | sudo tee /sys/class/shtrixmcmds/shtrixmcmd0/trigger
	Set brightness to 1 when ON
		echo 1 | sudo tee /sys/class/shtrixmcmds/shtrixmcmd0/brightness
	Set brightness to 0 when OFF
		echo 0 | sudo tee /sys/class/shtrixmcmds/shtrixmcmd0/brightness
	Get brightness status
		sudo cat /sys/class/shtrixmcmds/shtrixmcmd0/brightness

TODO:
[ ] catch errors in shtrixmcmdMessageHandler

*/
type shtrixmcmd struct {
	client       mqtt.Client
	debug        bool
	topic        string
	status       string
	shtrixmcmdID string
	deviceID     string
}

// a[len(a)-1:] last char

// newShtrixmcmd return new shtrixmcmd object.
func newShtrixmcmd(c mqtt.Client, topic string, debug bool) *shtrixmcmd {
	return &shtrixmcmd{
		client:       c,
		debug:        debug,
		topic:        topic,
		status:       "0",
		shtrixmcmdID: topic[len(topic)-1:],
		deviceID:     topic[len(topic)-3:],
	}
}

// Subscribe to topic
func (l *shtrixmcmd) Subscribe(qos byte) {

	go l.SrvMain()

	topic := l.topic + "/#"

	log.Println("[RUN] Subscribing:", qos, topic)

	if token := l.client.Subscribe(topic, qos, l.shtrixmcmdMessageHandler); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
	}
}

// UnSubscribe from topic
func (l *shtrixmcmd) UnSubscribe() {

	topic := l.topic + "/#"

	log.Println("[RUN] UnSubscribing:", topic)

	if token := l.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
	}
}

// PublishStatus Shtrixmcmd status
func (l *shtrixmcmd) PublishStatus(qos byte, deviceID, shtrixmcmdID string) {

	topic := l.topic + "/" + deviceID + "/" + shtrixmcmdID + "/status/shtrixmcmd"

	// publish result
	if token := l.client.Publish(topic, qos, false, l.status); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
	}

	// debug
	if l.debug {
		log.Println("[PUB]", qos, topic, l.status)
	}
}

// PublishPayload Relay status
func (l *shtrixmcmd) PublishPayload(qos byte, topicEnd string, payload string) {
	log.Println(" PUB Payload")

	topic := l.topic + "/" + topicEnd

	// publish result
	if token := l.client.Publish(topic, qos, false, payload); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
	}

	// // PublishStatus Shtrixmcmd status
	// func (l *shtrixmcmd) PublishStatusSensor(qos byte, deviceID, shtrixmcmdID string) {
	//
	// 	topic := l.topic + "/" + deviceID + "/" + shtrixmcmdID + "/status/sensor"
	//
	// 	// publish result
	// 	if token := l.client.Publish(topic, qos, false, l.status); token.Wait() && token.Error() != nil {
	// 		log.Println(token.Error())
	// 	}
	//
	// 	// debug
	// 	if l.debug {
	// 		log.Println("[PUB]", qos, topic, l.status)
	// 	}
	// }

	// // PublishStatus Shtrixmcmd status
	// func (l *shtrixmcmd) PublishADC(qos byte, deviceID, adcID string) {
	//
	// 	topic := l.topic + "/" + deviceID + "/" + adcID + "/status/adc"
	//
	// 	// publish result
	// 	if token := l.client.Publish(topic, qos, false, l.status); token.Wait() && token.Error() != nil {
	// 		log.Println(token.Error())
	// 	}
	//
	// 	// debug
	// 	if l.debug {
	// 		log.Println("[PUB]", qos, topic, l.status)
	// 	}
}

// shtrixmcmdMessageHandler set Shtrixmcmd to ON or OFF and get STATUS
func (l *shtrixmcmd) shtrixmcmdMessageHandler(client mqtt.Client, msg mqtt.Message) {

	// debug
	if l.debug {
		log.Println("[SUB]", msg.Qos(), msg.Topic(), string(msg.Payload()))
	}

	s1 := strings.Replace(msg.Topic(), l.topic, "", 1)
	s1 = strings.Replace(s1, "/", " ", -1)
	s_fields := strings.Fields(s1)
	device_id, _ := strconv.ParseUint(s_fields[0], 10, 64)
	shtrixmcmd_id, _ := strconv.ParseUint(s_fields[1], 10, 64)

	fmt.Printf(" len-sfields %v", len(s_fields))

	switch s_fields[len(s_fields)-1] {
	case "on":
		// receive message and DO
		switch string(msg.Payload()) {
		case "0":
			// logic when OFF
			log.Printf("%v %v", device_id, shtrixmcmd_id)
			// l.status = ShtrixmcmdOnOff(uint8(device_id), uint8(shtrixmcmd_id), 0)
			l.status = string(SendCmd([]byte{0x01}))
			log.Println("l.status", l.status)

			l.PublishStatus(0, s_fields[0], s_fields[1])
		case "1":
			// logic when ON
			log.Printf("%v %v", device_id, shtrixmcmd_id)

			// l.status = ShtrixmcmdOnOff(uint8(device_id), uint8(shtrixmcmd_id), 1)
			l.status = string(SendCmd([]byte{0x01}))
			log.Println("l.status", l.status)
			l.PublishStatus(0, s_fields[0], s_fields[1])
		case "3":
			// logic when ON

			log.Printf("%v %v", device_id, shtrixmcmd_id)
			// l.status = ShtrixmcmdWhile(uint8(device_id), uint8(shtrixmcmd_id), 3)
			l.status = string(SendCmd([]byte{0x01}))
			log.Println("l.status", l.status)
			l.PublishStatus(0, s_fields[0], s_fields[1])
		}

	case "sensor":
		// publish status
		// l.status = Status(uint8(device_id), uint8(shtrixmcmd_id))
		log.Printf("%v %v", device_id, shtrixmcmd_id)
		l.status = string(SendCmd([]byte{0x01}))
		log.Println("l.status", l.status)
		// l.PublishStatusSensor(0, s_fields[0], s_fields[1])

	case "setshtrixmcmddefaultmode":
		// receive message and DO
		switch string(msg.Payload()) {
		case "off":
			// publish status
			// l.status = SetConfig(uint8(device_id), uint8(shtrixmcmd_id), 2)
			log.Printf("%v %v", device_id, shtrixmcmd_id)
			l.status = string(SendCmd([]byte{0x01}))
			log.Println("l.status", l.status)
			l.PublishStatus(0, s_fields[0], s_fields[1])
		case "on":
			// publish status
			log.Printf("%v %v", device_id, shtrixmcmd_id)
			// l.status = SetConfig(uint8(device_id), uint8(shtrixmcmd_id), 1)
			l.status = string(SendCmd([]byte{0x01}))
			log.Println("l.status", l.status)
			l.PublishStatus(0, s_fields[0], s_fields[1])
			// case "blink":
			// 	// publish status
			// 	l.status = SetConfig(uint8(device_id), uint8(shtrixmcmd_id), 9)
			// 	log.Println("l.status", l.status)
			// 	l.PublishStatus(0, s_fields[0], s_fields[1])
			// case "pcn":
			// 	// publish status
			// 	l.status = SetConfig(uint8(device_id), uint8(shtrixmcmd_id), 10)
			// 	log.Println("l.status", l.status)
			// 	l.PublishStatus(0, s_fields[0], s_fields[1])
		}

		// case "setshtrixmcmdtime":
		// 	// receive message and DO
		// 	v, _ := strconv.ParseUint(string(msg.Payload()), 10, 64)
		// 	if v >= 1 && v <= 60 {
		// 		// l.status = SetConfig(uint8(device_id), uint8(shtrixmcmd_id)+4, uint8(v))
		// 		l.status = string(SendCmd([]byte{0x01}))
		// 		log.Println("l.status", l.status)
		// 		l.PublishStatus(0, s_fields[0], s_fields[1])
		// 	}
		//
		// case "changeaddress":
		// 	// receive message and DO
		// 	var newaddr uint8 = uint8(shtrixmcmd_id)
		// 	if newaddr >= 1 && newaddr <= 127 {
		// 		l.status = ChangeAddress(uint8(device_id), newaddr)
		// 		log.Println("l.status", l.status)
		// 		l.PublishStatus(0, s_fields[0], s_fields[1])
		// 	}

		// case "adc":
		// 	// publish status
		// 	l.status = ADC(uint8(device_id), uint8(shtrixmcmd_id))
		// 	log.Println("l.status", l.status)
		// 	l.PublishADC(0, s_fields[0], s_fields[1])
	}

}

// getBrightness
func getBrightness(shtrixmcmdID string) (string, error) {
	dat, err := ioutil.ReadFile("/sys/class/shtrixmcmds/shtrixmcmd" + shtrixmcmdID + "/brightness")
	if err != nil {
		return "", err
	}

	return strings.Trim(string(dat), "\r\n"), nil
}

// setBrightness
func setBrightness(shtrixmcmdID, data string) error {
	err := ioutil.WriteFile("/sys/class/shtrixmcmds/shtrixmcmd"+shtrixmcmdID+"/brightness", []byte(data), 0644)
	if err != nil {
		return err
	}
	return nil
}
