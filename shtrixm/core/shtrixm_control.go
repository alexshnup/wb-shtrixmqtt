package syscore

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/sigurn/crc8"
	"golang.org/x/text/encoding/charmap"
)

var (
	CRC8Shtrikh  = crc8.Params{0x31, 0xDE, false, false, 0x00, 0x80, "CRC-8/SHTRIKH"}
	ServerIpPort = "192.168.3.88:9999"
	// LocalIP      = "192.168.3.10"
	timeout            = 100
	Shtrixm_status_map = make(map[string]Shtrixm_status)
)
var icount byte

type Shtrixm_status struct {
	Ipaddr   string
	Status   byte
	Sn       string
	LastTime time.Time
}

type timeSlice []Shtrixm_status

//For Sorting Statuses by LastTime
func (p timeSlice) Len() int {
	return len(p)
}

//For Sorting Statuses by LastTime
func (p timeSlice) Less(i, j int) bool {
	return p[i].LastTime.Before(p[j].LastTime)
}

//For Sorting Statuses by LastTime
func (p timeSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func AddNewToMap(ip string, status byte) {
	// var Shtrixm_status_map = make(map[string]Shtrixm_status)
	Shtrixm_status_map[ip] = Shtrixm_status{Ipaddr: ip, LastTime: time.Now(), Status: status, Sn: "12345"}
}

func SortStatusesInMapByLastTime() {
	//Sort the map by LastTime
	date_sorted_reviews := make(timeSlice, 0, len(Shtrixm_status_map))
	for _, d := range Shtrixm_status_map {
		date_sorted_reviews = append(date_sorted_reviews, d)
	}
	fmt.Println(date_sorted_reviews)
	sort.Sort(date_sorted_reviews)
	fmt.Println(date_sorted_reviews)

}

//Crop only IP From String "192.168.80.11:9043"
func CropIpAddr(ipport string) string {
	x := strings.LastIndex(ipport, ":")
	return string([]byte(ipport)[:x])
}

func CRC8(buf []byte) []byte {

	table := crc8.MakeTable(CRC8Shtrikh)
	crc := crc8.Checksum(buf[1:], table)
	fmt.Printf("CRC-8: %X", crc)
	buf = append(buf, crc)
	return buf
}

//Error
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

//Get My IP
func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

// var out = []byte{0x02, 0x00, 0x01, 0x00, 0x01, 0x01} //addr 2 bytes

func SendResponse(conn *net.UDPConn, addr *net.UDPAddr, msgId, addr1, addr2 byte, cmd []byte) {
	var stx = []byte{0x02, 0x00, 0x01, 0x00, 0x00} //addr 1 bytes
	stx[1] = msgId
	stx[2] = addr1
	stx[3] = addr2
	stx[4] = byte(len(cmd))
	cmd = append(stx, cmd...)
	bufOut := CRC8(cmd) //append CRC
	// fmt.Printf("\nSend Reply: %+X\n", bufOut)
	fmt.Printf("\nSend Reply: \nstx:%X\tmsgId:%X\tAddr:%X %X\tLen:%X\tCmd:%X\tParams:%X\n", bufOut[0], bufOut[1], bufOut[2], bufOut[3], bufOut[4], bufOut[5], bufOut[6:])

	// fmt.Printf("\nSend Reply: \nstx:%+X\tmsgId:%X\tAddr:%X %X\tLen:%X\tCmd:%X\tParams:%X\tCRC8:%X\n", bufOut[0], bufOut[1], bufOut[2], bufOut[3], bufOut[4], bufOut[5], bufOut[6:len(bufOut)-1], bufOut[len(bufOut)])

	_, err := conn.WriteToUDP(bufOut, addr)
	if err != nil {
		fmt.Printf("Couldn't send response %v", err)
	}
}

func SendCmd(strHexByte []byte, ip_addr string) []byte {
	var bufOut []byte
	for i := icount; i <= 255; i++ {

		//addr 2 bytes
		if len(strHexByte) >= 1 {
			buf := []byte{0x02, 0x00, 0x00, 0x00, 0x00}
			buf[1] = byte(i)
			// buf[1] = byte(icount)
			buf[2] = 0x01
			buf[3] = 0x00
			buf[4] = byte(len(strHexByte))
			bufOut = append(buf, strHexByte...)
			bufOut = CRC8(bufOut)
		}

		// //addr 1 bytes
		// if len(strHexByte) >= 1 {
		// 	buf := []byte{0x02, 0x00, 0x00, 0x00}
		// 	buf[1] = byte(i)
		// 	// buf[1] = byte(icount)
		// 	buf[2] = addr1Uint8
		// 	buf[3] = byte(len(strHexByte))
		// 	bufOut = append(buf, strHexByte...)
		// 	bufOut = CRC8(bufOut)
		// }

		// fmt.Printf("\n bufOut: %+X\n", bufOut)
		// fmt.Printf("\n bufOut: %X\n", bufOut[0:4], , bufOut[5:6], )
		fmt.Printf("\nSend Reply: \nstx:%X\tmsgId:%X\tAddr:%X %X\tLen:%X\tCmd:%X\tParams:%X\n", bufOut[0], bufOut[1], bufOut[2], bufOut[3], bufOut[4], bufOut[5], bufOut[6:])

		p := make([]byte, 2048)
		// timeout := 100

		// conn, err := net.Dial("udp", ServerIpPort)
		conn, err := net.Dial("udp", ip_addr+":9999")
		// conn, err := net.Dial("udp",  "192.168.3.88:9999")
		// conn, err := net.DialTimeout("udp", "192.168.3.88:9999", time.Duration(timeout)*time.Millisecond)
		if err != nil {
			fmt.Printf("Some error %v", err)
			return []byte{0x00}
		}
		fmt.Fprintf(conn, string(bufOut))

		//The DeadLine set for read
		conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))

		n, err := bufio.NewReader(conn).Read(p)
		if err == nil {
			fmt.Printf("RX %d bytes \nstx:%+X\tmsgId:%X\tAddr:%X %X\tLen:%X\tCmd:%X\tParams:%X\tCRC8:%X\n", n, p[0], p[1], p[2], p[3], p[4], p[5], p[6:n-1], p[n])
			// fmt.Printf("RX %d bytes %+X\n", n, p[:n])
			icount = i
			return p[:n]
		} else {
			fmt.Printf("Some error %v\n", err)
		}
		conn.Close()
		// _, err = Conn.Write(bufOut)
		// if err != nil {
		// 	fmt.Println(bufOut, err)
		// } else {
		// 	fmt.Printf("\n%x\n", bufOut)
		// }

	}
	return []byte{0x00}
}

func (l *shtrixmcmd) SrvMain() {

	//Get My IP
	ip, err := externalIP()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("\nLocalIP: %s\n", ip)

	p := make([]byte, 255)
	addr := net.UDPAddr{
		Port: 9999,
		// IP:   net.ParseIP("192.168.3.10"),
		// IP: net.ParseIP(LocalIP),
		IP: net.ParseIP(ip),
	}
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}

	for {
		n, remoteaddr, err := ser.ReadFromUDP(p)
		// fmt.Printf("Read from %v %d bytes %x \n", remoteaddr, n, p[:n])
		// fmt.Printf("Read from %v %d bytes \nstx:%X\tmsgId:%X\tAddr:%X %X\tLen:%X\tCmd:%X\tParams:%X\tCRC8:%X\n", remoteaddr, n, p[0], p[1], p[2], p[3], p[4], p[5], p[6:n-1], p[n])
		fmt.Printf("Read from %v %d bytes \nstx:%X\tmsgId:%X\tAddr:%X %X\tLen:%X\tCmd:%X\tParams:%X\n", remoteaddr, n, p[0], p[1], p[2], p[3], p[4], p[5], p[6:])

		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}
		msgId := p[1]
		addr1 := p[2]
		addr2 := p[3]

		cmd := []byte{0x70, 0x00}

		switch p[6] {
		case 0x14:
			cmd = []byte{0x70, 0x00}
			go SendResponse(ser, remoteaddr, msgId, addr1, addr2, cmd)
			//Save to map
			AddNewToMap(CropIpAddr(remoteaddr.String()), p[22])
			//Sort and Print MAP fo test
			SortStatusesInMapByLastTime()

			//Publish Status
			// l.PublishPayload(0, CropIpAddr(remoteaddr.String())+"/"+fmt.Sprintf("%X", p[7])+"/"+fmt.Sprintf("%X", p[8])+"/"+fmt.Sprintf("%X", p[9:17])+"/status/", fmt.Sprintf("%X", p[22]))
			// l.PublishPayload(0, CropIpAddr(remoteaddr.String())+"/"+fmt.Sprintf("%d", p[9:17])+"/status/", fmt.Sprintf("%X", p[22]))
			l.PublishPayload(0, CropIpAddr(remoteaddr.String())+"/status/", fmt.Sprintf("%X", p[22]))

		case 0x81:
			cmd = []byte{0x70, 0x00}
			SendResponse(ser, remoteaddr, msgId, addr1, addr2, cmd)

			l.PublishPayload(0, CropIpAddr(remoteaddr.String())+"/"+fmt.Sprintf("%X", p[9:17])+"/inquiry/", fmt.Sprintf("%X", p[10:n-5]))

			//p[10:n-5]
		}
		// go sendResponse(ser, remoteaddr, msgId, addr1, addr2, cmd)

		// go sendResponse(ser, remoteaddr, msgId, addr1, addr2, []byte{0x70, 0x00})

	}
}

func DecodeWindows1251(ba []uint8) []uint8 {
	dec := charmap.Windows1251.NewDecoder()
	out, _ := dec.Bytes(ba)
	return out
}

func EncodeWindows1251(ba []uint8) []uint8 {
	enc := charmap.Windows1251.NewEncoder()
	out, _ := enc.String(string(ba))
	return []uint8(out)
}

// package syscore
//
// import (
// 	"fmt"
// 	"log"
// 	"strconv"
// 	"time"
//
// 	tarmserial "github.com/alexshnup/serial"
// 	"github.com/alexshnup/wb-shtrixmqtt/conf"
// )
//
// // CRC 8-bit Calculate method Dallas 1-wire with prepared tables
// func crc8dallas(Command1 []byte) []byte {
// 	// CRCTable := [256]byte{0, 94, 188, 226, 97, 63, 221, 131, 194, 156, 126, 32, 163, 253, 31, 65, 157, 195, 33, 127, 252, 162, 64, 30, 95, 1, 227, 189, 62, 96, 130, 220, 35, 125, 159, 193, 66, 28, 254, 160, 225, 191, 93, 3, 128, 222, 60, 98, 190, 224, 2, 92, 223, 129, 99, 61, 124, 34, 192, 158, 29, 67, 161, 255, 70, 24, 250, 164, 39, 121, 155, 197, 132, 218, 56, 102, 229, 187, 89, 7, 219, 133, 103, 57, 186, 228, 6, 88, 25, 71, 165, 251, 120, 38, 196, 154, 101, 59, 217, 135, 4, 90, 184, 230, 167, 249, 27, 69, 198, 152, 122, 36, 248, 166, 68, 26, 153, 199, 37, 123, 58, 100, 134, 216, 91, 5, 231, 185, 140, 210, 48, 110, 237, 179, 81, 15, 78, 16, 242, 172, 47, 113, 147, 205, 17, 79, 173, 243, 112, 46, 204, 146, 211, 141, 111, 49, 178, 236, 14, 80, 175, 241, 19, 77, 206, 144, 114, 44, 109, 51, 209, 143, 12, 82, 176, 238, 50, 108, 142, 208, 83, 13, 239, 177, 240, 174, 76, 18, 145, 207, 45, 115, 202, 148, 118, 40, 171, 245, 23, 73, 8, 86, 180, 234, 105, 55, 213, 139, 87, 9, 235, 181, 54, 104, 138, 212, 149, 203, 41, 119, 244, 170, 72, 22, 233, 183, 85, 11, 136, 214, 52, 106, 43, 117, 151, 201, 74, 20, 246, 168, 116, 42, 200, 150, 21, 75, 169, 247, 182, 232, 10, 84, 215, 137, 107, 53}
// 	CRCTable := [256]byte{0x00, 0x5E, 0x0BC, 0x0E2, 0x61, 0x3F, 0x0DD, 0x83, 0x0C2, 0x9C, 0x7E, 0x20, 0x0A3, 0x0FD, 0x1F, 0x41, 0x9D, 0x0C3, 0x21, 0x7F, 0x0FC, 0x0A2, 0x40, 0x1E, 0x5F, 0x01, 0x0E3, 0x0BD, 0x3E, 0x60, 0x82, 0x0DC, 0x23, 0x7D, 0x9F, 0x0C1, 0x42, 0x1C, 0x0FE, 0x0A0, 0x0E1, 0x0BF, 0x5D, 0x03, 0x80, 0x0DE, 0x3C, 0x62, 0x0BE, 0x0E0, 0x02, 0x5C, 0x0DF, 0x81, 0x63, 0x3D, 0x7C, 0x22, 0x0C0, 0x9E, 0x1D, 0x43, 0x0A1, 0x0FF, 0x46, 0x18, 0x0FA, 0x0A4, 0x27, 0x79, 0x9B, 0x0C5, 0x84, 0x0DA, 0x38, 0x66, 0x0E5, 0x0BB, 0x59, 0x07, 0x0DB, 0x85, 0x67, 0x39, 0x0BA, 0x0E4, 0x06, 0x58, 0x19, 0x47, 0x0A5, 0x0FB, 0x78, 0x26, 0x0C4, 0x9A, 0x65, 0x3B, 0x0D9, 0x87, 0x04, 0x5A, 0x0B8, 0x0E6, 0x0A7, 0x0F9, 0x1B, 0x45, 0x0C6, 0x98, 0x7A, 0x24, 0x0F8, 0x0A6, 0x44, 0x1A, 0x99, 0x0C7, 0x25, 0x7B, 0x3A, 0x64, 0x86, 0x0D8, 0x5B, 0x05, 0x0E7, 0x0B9, 0x8C, 0x0D2, 0x30, 0x6E, 0x0ED, 0x0B3, 0x51, 0x0F, 0x4E, 0x10, 0x0F2, 0x0AC, 0x2F, 0x71, 0x93, 0x0CD, 0x11, 0x4F, 0x0AD, 0x0F3, 0x70, 0x2E, 0x0CC, 0x92, 0x0D3, 0x8D, 0x6F, 0x31, 0x0B2, 0x0EC, 0x0E, 0x50, 0x0AF, 0x0F1, 0x13, 0x4D, 0x0CE, 0x90, 0x72, 0x2C, 0x6D, 0x33, 0x0D1, 0x8F, 0x0C, 0x52, 0x0B0, 0x0EE, 0x32, 0x6C, 0x8E, 0x0D0, 0x53, 0x0D, 0x0EF, 0x0B1, 0x0F0, 0x0AE, 0x4C, 0x12, 0x91, 0x0CF, 0x2D, 0x73, 0x0CA, 0x94, 0x76, 0x28, 0x0AB, 0x0F5, 0x17, 0x49, 0x08, 0x56, 0x0B4, 0x0EA, 0x69, 0x37, 0x0D5, 0x8B, 0x57, 0x09, 0x0EB, 0x0B5, 0x36, 0x68, 0x8A, 0x0D4, 0x95, 0x0CB, 0x29, 0x77, 0x0F4, 0x0AA, 0x48, 0x16, 0x0E9, 0x0B7, 0x55, 0x0B, 0x88, 0x0D6, 0x34, 0x6A, 0x2B, 0x75, 0x97, 0x0C9, 0x4A, 0x14, 0x0F6, 0x0A8, 0x74, 0x2A, 0x0C8, 0x96, 0x15, 0x4B, 0x0A9, 0x0F7, 0x0B6, 0x0FC, 0x0A, 0x54, 0x0D7, 0x89, 0x6B, 0x35}
//
// 	size := len(Command1) - 1
// 	Command1[size] = 0
// 	for i := 0; i <= int(size)-1; i++ {
// 		Command1[size] = CRCTable[Command1[i]^Command1[size]]
// 	}
// 	return Command1
// }
//
// //Send and Receive Answer bytes
// func write_serial(send []byte) []byte {
// 	delay := 50 * time.Millisecond
//
// 	//c := &serial.Config{Name: "/dev/ttyS2", Baud: 9600, ReadTimeout: time.Millisecond * 500}
// 	// c := &tarmserial.Config{Name: "/dev/ttyAPP1", Baud: 9600, ReadTimeout: time.Millisecond * 5000}
// 	c := &tarmserial.Config{Name: conf.Config.Serial.Port, Baud: conf.Config.Serial.Baud, ReadTimeout: time.Millisecond * conf.Config.Serial.ReadTimeout}
//
// 	//c := new(serial.Config)
// 	//c.Name = "/dev/ttyAPP1"
// 	//c.Name = "/dev/ttyS2"
// 	//c.Baud = 9600
// 	//c.ReadTimeout = time.Millisecond * 5000
// 	//c.Size = 8
// 	//c.StopBits = 0
//
// 	s, err := tarmserial.OpenPort(c)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	//Send Data
// 	//n, err := s.Write([]byte("test"))
// 	//s.Flush()
// 	time.Sleep(delay)
// 	_, err = s.Write(crc8dallas(send))
// 	if err != nil {
// 		log.Printf("Error Oper Port...")
// 		log.Fatal(err)
// 	}
//
// 	//Need delay for correct the receive answer
// 	time.Sleep(delay)
//
// 	//Receive Respond
// 	buf := make([]byte, 6)
// 	// _, err = s.Read(buf)
// 	//nr, err_read := s.Read(buf)
// 	if _, err = s.Read(buf); err != nil {
// 		buf = []byte{00, 00, 00, 00, 00, 00}
// 		log.Printf("Error Read...")
// 		// log.Fatal(err)
// 	}
//
// 	//Discards data written to the port but not transmitted,
// 	//or data received but not read
// 	s.Flush()
//
// 	//Close Serial Port
// 	s.Close()
//
// 	return []byte(buf)
// }
//
// func status_serial(result []byte, typeStatus uint8, channel uint8) (uint8, string) {
// 	fmt.Printf("%x\n", result)
// 	switch typeStatus {
// 	case 0:
// 		if rune(result[4]) == 0x18 {
// 			log.Printf("Enter - shtrixmcmd %d - OK", channel)
// 			return channel, "enter_ok"
// 		}
// 	case 1:
// 		if rune(result[2]) == 0x42 {
// 			log.Printf("Command OK - shtrixmcmd %d", channel)
// 			return channel, "shtrixmcmd_ok"
// 		}
// 	case 2:
// 		if rune(result[4]) == 0x01 && (rune(result[3]) == rune(channel)) {
// 			log.Printf("Port %d - ON", channel)
// 			return channel, "on"
// 		} else if rune(result[4]) == 0x00 && (rune(result[3]) == rune(channel)) {
// 			log.Printf("Port %d - OFF", channel)
// 			return channel, "off"
// 		}
// 	case 3:
// 		if rune(result[4]) == 0x03 && (rune(result[3]) == rune(channel)) {
// 			log.Printf("Activate the shtrixmcmd %d for a time", channel)
// 			return channel, "while"
// 		}
//
// 	case 4:
// 		return channel, fmt.Sprintf("%d", result[4])
//
// 	case 5:
// 		return channel, fmt.Sprintf("%d", result[3])
//
// 	case 6:
// 		return channel, fmt.Sprintf("%d", result[3])
//
// 	case 7:
// 		if rune(result[5]) == 0x95 {
// 			log.Printf("Status %d - Open", channel)
// 			return channel, "open"
// 		} else if rune(result[5]) == 0x98 {
// 			log.Printf("Status %d - Close", channel)
// 			return channel, "close"
// 		}
//
// 	}
// 	return channel, "none"
// }
//
// func ProgramDefaultStateShtrixmcmd_ON(addr uint8, shtrixmcmd uint8) string {
// 	Command1 := []byte{127, 8, 0, 65, 1, 0, 0, 1, 0}
// 	CommandSave := []byte{127, 6, 0, 23, 0, 0, 0}
// 	Command1[0] = addr
// 	Command1[4] = shtrixmcmd
// 	CommandSave[0] = addr
// 	_, out := status_serial(write_serial(crc8dallas(Command1)), 1, shtrixmcmd)
// 	_, out = status_serial(write_serial(crc8dallas(CommandSave)), 0, shtrixmcmd)
// 	return out
// }
//
// func ProgramDefaultStateShtrixmcmd_OFF(addr uint8, shtrixmcmd uint8) string {
// 	Command1 := []byte{127, 8, 0, 65, 1, 0, 0, 2, 0}
// 	// Command1 := conf.Config.Bolid.ShtrixmcmdOFF
// 	CommandSave := []byte{127, 6, 0, 23, 0, 0, 0}
// 	Command1[0] = addr
// 	Command1[4] = shtrixmcmd
// 	CommandSave[0] = addr
// 	_, out := status_serial(write_serial(crc8dallas(Command1)), 1, shtrixmcmd)
// 	_, out = status_serial(write_serial(crc8dallas(CommandSave)), 0, shtrixmcmd)
// 	return out
// }
//
// // func StatusShtrixmcmd(addr uint8, shtrixmcmd uint8) string {
// // 	Command1 := []byte{127, 8, 0, 67, 1, 0, 0, 1, 0}
// // 	Command1[0] = addr
// // 	Command1[4] = shtrixmcmd
// // 	_, out := status_serial(write_serial(crc8dallas(Command1)), 2, shtrixmcmd)
// // 	return out
// // }
//
// func Status(addr uint8, shtrixmcmd uint8) string {
// 	Command1 := []byte{127, 0x06, 0x00, 0x19, 0x01, 0x00, 0xFF}
// 	Command1[0] = addr
// 	Command1[4] = shtrixmcmd
// 	_, out := status_serial(write_serial(crc8dallas(Command1)), 7, shtrixmcmd)
// 	return out
// }
//
// func ShtrixmcmdOnOff(addr uint8, shtrixmcmd uint8, on uint8) string {
// 	Command1 := []byte{127, 0x06, 0x00, 0x15, 0x01, 0x01, 0xFF}
// 	Command1[0] = addr
// 	Command1[4] = shtrixmcmd
// 	Command1[5] = on // 0-off 1-on 3-blink ...
// 	_, out := status_serial(write_serial(crc8dallas(Command1)), 2, shtrixmcmd)
// 	return out
// }
//
// func ShtrixmcmdWhile(addr uint8, shtrixmcmd uint8, on uint8) string {
// 	Command1 := []byte{127, 0x06, 0x00, 0x15, 0x01, 0x01, 0xFF}
// 	Command1[0] = addr
// 	Command1[4] = shtrixmcmd
// 	Command1[5] = on // 0-off 1-on 3-blink ...
// 	_, out := status_serial(write_serial(crc8dallas(Command1)), 3, shtrixmcmd)
// 	return out
// }
//
// func ADC(addr uint8, input uint8) string {
// 	Command1 := []byte{127, 0x06, 0x00, 0x1B, 0x01, 0x01, 0xFF}
// 	Command1[0] = addr
// 	Command1[4] = input
// 	_, out := status_serial(write_serial(crc8dallas(Command1)), 4, input)
// 	voltageInPopugai, _ := strconv.ParseUint(out, 10, 8)
// 	out = strconv.FormatFloat((float64(voltageInPopugai)*134)/1000, 'f', 1, 32)
// 	return out
// }
//
// func SetConfig(addr, k, v uint8) string {
// 	//set mode
// 	Command1 := []byte{127, 0x08, 0x00, 0x41, 0x01, 0x00, 0x00, 0x02, 0xFF}
//
// 	Command1[0] = addr
// 	Command1[4] = k
// 	Command1[7] = v
// 	_, out := status_serial(write_serial(crc8dallas(Command1)), 5, k)
// 	return out
// }
//
// func ChangeAddress(oldaddr, newaddr uint8) string {
// 	//set mode
// 	Command1 := []byte{127, 0x06, 0x00, 0x0f, 0x01, 0x01, 0xFF}
//
// 	Command1[0] = oldaddr
// 	Command1[4] = newaddr
// 	Command1[5] = newaddr
// 	_, out := status_serial(write_serial(crc8dallas(Command1)), 6, newaddr)
// 	return out
// }
//
// // func Shtrixmcmd_OFF(addr uint8, shtrixmcmd uint8) string {
// // 	Command1 := []byte{127, 0x06, 0x00, 0x15, 0x01, 0x00, 0xFF}
// // 	Command1[0] = addr
// // 	Command1[4] = shtrixmcmd
// // 	_, out := status_serial(write_serial(crc8dallas(Command1)), 1, shtrixmcmd)
// // 	return out
// // }
