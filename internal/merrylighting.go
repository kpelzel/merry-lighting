package internal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v3"
	"tinygo.org/x/bluetooth"
)

type merryLighting struct {
	colorChan chan color
}

var adapter = bluetooth.DefaultAdapter

func StartMerryLighting(debug bool, config string) error {
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	// read config file
	confBytes, err := ioutil.ReadFile(config)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	conf := &conf{}
	err = yaml.Unmarshal(confBytes, conf)
	if err != nil {
		log.Fatalf("failed to unmarshal config file: %v", err)
	}

	log.Debugf("config: %+v", conf)

	// Enable BLE interface.
	err = adapter.Enable()
	if err != nil {
		log.Fatalf("failed to enable ble stack: %v", err)
	}

	// connect to bluetooth lights
	devs, err := connectToLights(conf.Output)
	if err != nil {
		log.Fatalf("failed to connect to lights: %v", err)
	}

	// disconnect from lights when program ends
	defer func() {
		for _, d := range devs {
			d.Disconnect()
		}
	}()

	// get characteristics
	chars, err := getCharacteristics(devs)
	if err != nil {
		log.Fatalf("failed to get characteristics: %v", err)
	}

	// turn on lights
	for cn, c := range chars {
		err := on(c)
		if err != nil {
			log.Fatalf("failed to turn on dev[%v]: %v", cn, err)
		}
	}

	ml := &merryLighting{
		colorChan: make(chan color, 1000),
	}

	//  monitor for color changes
	go ml.listenForColor()

	addr := net.JoinHostPort(conf.Input.IP.String(), strconv.Itoa(conf.Input.Port))
	// listen for incoming sACN packets
	udpServer, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer udpServer.Close()

	prevValue := []byte{}

	log.Infof("listening on %v for sACN packets", addr)
	for {
		buf := make([]byte, 1024)
		_, _, err := udpServer.ReadFrom(buf)
		if err != nil {
			log.Errorf("error reading udp packet: %v", err)
		}
		// logger.Printf("addr: %v", addr)
		// logger.Printf("bytes read: %v", n)
		// logger.Printf("buf: %v", buf)
		// logger.Printf("buf string: %v", string(buf))

		p, err := DecodePacket(buf)
		if err != nil {
			log.Errorf("error parsing packet: %v", err)
		} else if int(binary.BigEndian.Uint16(p.frame.Universe)) == conf.Input.Universe {
			// logger.Printf("root: %+v", p.root)
			// logger.Printf("frame: %+v", p.frame)
			// logger.Printf("dmp: %+v", p.dmp)

			log.Debugf("received sACN packet:")
			log.Debugf("universe: %v", binary.BigEndian.Uint16(p.frame.Universe))
			log.Debugf("source name: %s", string(p.frame.SourceName))
			log.Debugf("value count: %v", binary.BigEndian.Uint16(p.dmp.PropertyValCount))
			log.Debugf("value: %v", p.dmp.PropertyVal)

			if !bytes.Equal(prevValue, p.dmp.PropertyVal) {
				log.Debugf("new packet different than previous: %v vs %v", prevValue, p.dmp.PropertyVal)
				prevValue = p.dmp.PropertyVal
				for ln, c := range chars {
					rb := p.dmp.PropertyVal[conf.Output[ln].RedByte]
					gb := p.dmp.PropertyVal[conf.Output[ln].GreenByte]
					bb := p.dmp.PropertyVal[conf.Output[ln].BlueByte]

					log.Debugf("sending red: %v", rb)
					log.Debugf("sending green: %v", gb)
					log.Debugf("sending blue: %v", bb)

					select {
					case ml.colorChan <- color{char: c, Red: rb, Green: gb, Blue: bb}:
					default:
						log.Debug("bluetooth busy, color not sent")
					}

					// err := setColor(c, rb, gb, bb)
					// if err != nil {
					// 	log.Errorf("failed to set color for dev[%v]: %v", ln, err)
					// }
				}
			}
		}
	}
}

func DecodePacket(data []byte) (*e131Packet, error) {
	if len(data) < 125 {
		return nil, fmt.Errorf("invalid packet size: %v", len(data))
	}
	p := &e131Packet{
		root: &e131PacketRoot{
			PreambleSize:  data[0:2],
			PostambleSize: data[2:4],
			PacketID:      data[4:16],
			FlagsLength:   data[16:18],
			Vector:        data[18:22],
			CID:           data[22:38],
		},
		frame: &e131PacketFrame{
			FlagsLength: data[38:40],
			Vector:      data[40:44],
			SourceName:  data[44:108],
			Priority:    data[108],
			SyncAddr:    data[109:111],
			SeqNum:      data[111],
			Options:     data[112],
			Universe:    data[113:115],
		},
		dmp: &e131PacketDMP{
			FlagsLength:       data[115:117],
			Vector:            data[117],
			AddrType:          data[118],
			FirstPropertyAddr: data[119:121],
			AddrInc:           data[121:123],
			PropertyValCount:  data[123:125],
			PropertyVal:       data[125:],
		},
	}

	return p, nil
}

func EncodePacket(p *e131Packet) []byte {
	return concatMultipleSlices([][]byte{
		p.root.PreambleSize,
		p.root.PostambleSize,
		p.root.PacketID,
		p.root.FlagsLength,
		p.root.Vector,
		p.root.CID,
		p.frame.FlagsLength,
		p.frame.Vector,
		p.frame.SourceName,
		[]byte{p.frame.Priority},
		p.frame.SyncAddr,
		[]byte{p.frame.SeqNum},
		[]byte{p.frame.Options},
		p.frame.Universe,
		p.dmp.FlagsLength,
		[]byte{p.dmp.Vector},
		[]byte{p.dmp.AddrType},
		p.dmp.FirstPropertyAddr,
		p.dmp.AddrInc,
		p.dmp.PropertyValCount,
		p.dmp.PropertyVal,
	})
}

func concatMultipleSlices(slices [][]byte) []byte {
	var totalLen int

	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]byte, totalLen)

	var i int

	for _, s := range slices {
		i += copy(result[i:], s)
	}

	return result
}

func getCharacteristics(devs map[string]*bluetooth.Device) (map[string]*bluetooth.DeviceCharacteristic, error) {
	finalCharacteristics := make(map[string]*bluetooth.DeviceCharacteristic)
	serWID := bluetooth.New16BitUUID(0xFFD5)
	charWID := bluetooth.New16BitUUID(0xFFD9)
	serRID := bluetooth.New16BitUUID(0xFFD0)
	// charRID := bluetooth.New16BitUUID(0xFFD4)

	for dn, dev := range devs {
		log.Debugf("looking for services: %v %v", serWID, serRID)
		ser, err := dev.DiscoverServices([]bluetooth.UUID{serWID, serRID})
		if err != nil {
			return nil, fmt.Errorf("failed to discover services for dev[%v]: %v", dn, err)
		}
		// fmt.Printf("found services: %+v\n", ser)
		if len(ser) < 2 {
			return nil, fmt.Errorf("failed to discover enough services for dev[%v]: %v", dn, len(ser))
		}

		// // read for notification
		// rChars, err := ser[1].DiscoverCharacteristics([]bluetooth.UUID{charRID})
		// if err != nil {
		// 	fmt.Printf("failed to discover read characteristic: %v", err)
		// 	return
		// }
		// fmt.Printf("found read characteristics: %+v\n", rChars)
		// if len(rChars) < 1 {
		// 	fmt.Printf("failed to discover enough characteristics: %v\n", len(rChars))
		// 	return
		// }
		// rChars[0].EnableNotifications(func(buff []byte) { fmt.Printf("got notification: %v", buff) })

		// get characteristic
		wChars, err := ser[0].DiscoverCharacteristics([]bluetooth.UUID{charWID})
		if err != nil {
			return nil, fmt.Errorf("failed to discover write characteristic for dev[%v]: %v", dn, err)
		}
		// fmt.Printf("found write characteristics: %+v\n", wChars)
		if len(wChars) < 1 {
			return nil, fmt.Errorf("failed to discover enough characteristics for dev[%v]: %v", dn, len(wChars))
		}

		finalCharacteristics[dn] = &wChars[0]
	}

	return finalCharacteristics, nil
}
