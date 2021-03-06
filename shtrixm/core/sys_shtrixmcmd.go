package syscore

import (
	"encoding/hex"
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
	reply        string
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
		reply:        "0",
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
	if token := l.client.Publish(topic, qos, false, l.reply); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
	}

	// debug
	if l.debug {
		log.Println("[PUB]", qos, topic, l.reply)
	}
}

func (l *shtrixmcmd) PublishReply(qos byte, deviceID, shtrixmcmdID string) {

	topic := l.topic + "/" + deviceID + "/" + shtrixmcmdID
	// publish result
	if token := l.client.Publish(topic, qos, false, fmt.Sprintf("%X", l.reply)); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
	}

	// debug
	if l.debug {
		log.Println("[PUB]", qos, topic, l.reply)
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
	// 	if token := l.client.Publish(topic, qos, false, l.reply); token.Wait() && token.Error() != nil {
	// 		log.Println(token.Error())
	// 	}
	//
	// 	// debug
	// 	if l.debug {
	// 		log.Println("[PUB]", qos, topic, l.reply)
	// 	}
	// }

	// // PublishStatus Shtrixmcmd status
	// func (l *shtrixmcmd) PublishADC(qos byte, deviceID, adcID string) {
	//
	// 	topic := l.topic + "/" + deviceID + "/" + adcID + "/status/adc"
	//
	// 	// publish result
	// 	if token := l.client.Publish(topic, qos, false, l.reply); token.Wait() && token.Error() != nil {
	// 		log.Println(token.Error())
	// 	}
	//
	// 	// debug
	// 	if l.debug {
	// 		log.Println("[PUB]", qos, topic, l.reply)
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

	fmt.Printf("\n sfields %v len-sfields %v \n", s_fields, len(s_fields))

	switch s_fields[len(s_fields)-1] {
	case "set":
		switch string(msg.Payload()) {
		case "open":
			log.Printf("open______0______%v %v", s_fields[0], s_fields[1])
			// l.reply = ShtrixmcmdOnOff(uint8(device_id), uint8(shtrixmcmd_id), 0)
			l.reply = string(SendCmd([]byte{0x80, 0x02}, string(s_fields[0])))
			log.Println("l.reply", l.reply)
			l.PublishReply(0, s_fields[0], "reply")
		case "close":
			log.Printf("close______0______%v %v", s_fields[0], s_fields[1])
			// l.reply = ShtrixmcmdOnOff(uint8(device_id), uint8(shtrixmcmd_id), 0)
			l.reply = string(SendCmd([]byte{0x80, 0x00}, string(s_fields[0])))
			log.Println("l.reply", l.reply)
			l.PublishReply(0, s_fields[0], "reply")
		}

	case "get":
		switch string(msg.Payload()) {
		case "devicetype":
			log.Printf("devicetype______0______%v %v", s_fields[0], s_fields[1])
			// l.reply = ShtrixmcmdOnOff(uint8(device_id), uint8(shtrixmcmd_id), 0)
			l.reply = string(SendCmd([]byte{0x01}, string(s_fields[0])))
			log.Println("l.reply", l.reply)
			l.PublishReply(0, s_fields[0], "reply")
		case "params":
			log.Printf("params______0______%v %v", s_fields[0], s_fields[1])
			// l.reply = ShtrixmcmdOnOff(uint8(device_id), uint8(shtrixmcmd_id), 0)
			l.reply = string(SendCmd([]byte{0x02}, string(s_fields[0])))
			log.Println("l.reply", l.reply)
			l.PublishReply(0, s_fields[0], "reply")
		}

	case "cmd":
		text2 := string(msg.Payload())
		text2 = strings.Replace(text2, " ", "", -1)
		text2 = strings.Trim(text2, "\n")

		strHexByte, err := hex.DecodeString(text2)
		CheckError(err)

		log.Printf("cmd______0______%v %v", s_fields[0], s_fields[1])
		// l.reply = ShtrixmcmdOnOff(uint8(device_id), uint8(shtrixmcmd_id), 0)
		l.reply = string(SendCmd(strHexByte, string(s_fields[0])))
		log.Println("l.reply", l.reply)
		l.PublishReply(0, s_fields[0], "reply_cmd")

	case "allow":
		log.Printf("allow______0______%v %v %v", s_fields[0], s_fields[1], string(msg.Payload()))
		//scanned number
		txt := string(s_fields[1])
		txt = strings.Replace(txt, " ", "", -1)
		txt = strings.Trim(txt, "\n")

		strHexByte, err := hex.DecodeString(txt)
		CheckError(err)

		cmd := []byte{0x73, 0x77}
		//cmd = append(cmd, p[10:n-5]...)
		cmd = append(cmd, strHexByte...)
		//AccessAllow
		cmdEndOK := []byte{0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

		cmd = append(cmd, cmdEndOK...)

		// text = []byte("                         ПРОХОДИТЕ                         ")
		text1 := []byte("                    ")
		text2 := []byte("     ПРОХОДИТЕ      ")
		text3 := []byte("                    ")
		text4 := []byte("                    ")
		text := append(text1, text2...)
		text = append(text, text3...)
		text = append(text, text4...)
		win1251 := EncodeWindows1251(text)
		cmd = append(cmd, win1251...)

		result := SendCmd(cmd, string(s_fields[0]))
		fmt.Printf("\nresult____%v\n", result)

	case "deny":
		log.Printf("allow______0______%v %v %v", s_fields[0], s_fields[1], string(msg.Payload()))
		//scanned number
		txt := string(s_fields[1])
		txt = strings.Replace(txt, " ", "", -1)
		txt = strings.Trim(txt, "\n")

		strHexByte, err := hex.DecodeString(txt)
		CheckError(err)

		cmd := []byte{0x73, 0x77}
		cmd = append(cmd, strHexByte...)
		//AccessDeny
		fmt.Print("\n0x34\n")
		cmdEndOK := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

		cmd = append(cmd, cmdEndOK...)

		text1 := []byte("                    ")
		text2 := []byte("  ДОСТУП ЗАПРЕЩЕН   ")
		text3 := []byte("                    ")
		text4 := []byte("                    ")
		text := append(text1, text2...)
		text = append(text, text3...)
		text = append(text, text4...)
		win1251 := EncodeWindows1251(text)
		cmd = append(cmd, win1251...)

		result := SendCmd(cmd, string(s_fields[0]))
		fmt.Printf("\nresult____%v\n", result)

	case "setshtrixmcmddefaultmode":
		// receive message and DO
		switch string(msg.Payload()) {
		case "off":
			// publish status
			// l.reply = SetConfig(uint8(device_id), uint8(shtrixmcmd_id), 2)
			log.Printf("%v %v", device_id, shtrixmcmd_id)
			l.reply = string(SendCmd([]byte{0x01}, string(s_fields[0])))
			log.Println("l.reply", l.reply)
			l.PublishStatus(0, s_fields[0], s_fields[1])
		case "on":
			// publish status
			log.Printf("%v %v", device_id, shtrixmcmd_id)
			// l.reply = SetConfig(uint8(device_id), uint8(shtrixmcmd_id), 1)
			l.reply = string(SendCmd([]byte{0x01}, string(s_fields[0])))
			log.Println("l.reply", l.reply)
			l.PublishStatus(0, s_fields[0], s_fields[1])
			// case "blink":
			// 	// publish status
			// 	l.reply = SetConfig(uint8(device_id), uint8(shtrixmcmd_id), 9)
			// 	log.Println("l.reply", l.reply)
			// 	l.PublishStatus(0, s_fields[0], s_fields[1])
			// case "pcn":
			// 	// publish status
			// 	l.reply = SetConfig(uint8(device_id), uint8(shtrixmcmd_id), 10)
			// 	log.Println("l.reply", l.reply)
			// 	l.PublishStatus(0, s_fields[0], s_fields[1])
		}

		// case "setshtrixmcmdtime":
		// 	// receive message and DO
		// 	v, _ := strconv.ParseUint(string(msg.Payload()), 10, 64)
		// 	if v >= 1 && v <= 60 {
		// 		// l.reply = SetConfig(uint8(device_id), uint8(shtrixmcmd_id)+4, uint8(v))
		// 		l.reply = string(SendCmd([]byte{0x01}))
		// 		log.Println("l.reply", l.reply)
		// 		l.PublishStatus(0, s_fields[0], s_fields[1])
		// 	}
		//
		// case "changeaddress":
		// 	// receive message and DO
		// 	var newaddr uint8 = uint8(shtrixmcmd_id)
		// 	if newaddr >= 1 && newaddr <= 127 {
		// 		l.reply = ChangeAddress(uint8(device_id), newaddr)
		// 		log.Println("l.reply", l.reply)
		// 		l.PublishStatus(0, s_fields[0], s_fields[1])
		// 	}

		// case "adc":
		// 	// publish status
		// 	l.reply = ADC(uint8(device_id), uint8(shtrixmcmd_id))
		// 	log.Println("l.reply", l.reply)
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
