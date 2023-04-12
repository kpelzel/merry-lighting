package internal

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

func connectToLights(lights map[string]confLight) (map[string]*bluetooth.Device, error) {
	finalDevs := make(map[string]*bluetooth.Device)

	for ln, l := range lights {
		mac, err := bluetooth.ParseMAC(l.MACAddress)
		if err != nil {
			for _, d := range finalDevs {
				d.Disconnect()
			}
			return nil, fmt.Errorf("failed to parse mac address[%v]: %v", l.MACAddress, err)
		}

		address := bluetooth.MACAddress{
			MAC: mac,
		}

		log.Infof("connecting to light[%v] at %v...", ln, l.UUID)
		dev, err := adapter.Connect(address, bluetooth.ConnectionParams{})
		if err != nil {
			for _, d := range finalDevs {
				d.Disconnect()
			}
			return nil, fmt.Errorf("failed to connect to device[%v]: %v", ln, err)
		}
		log.Infof("successfully connected to light[%v] at %v", ln, l.UUID)

		finalDevs[ln] = dev
	}

	return finalDevs, nil
}
