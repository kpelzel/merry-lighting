package internal

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"tinygo.org/x/bluetooth"
)

func connectToLights(lights map[string]confLight) (map[string]*bluetooth.Device, error) {
	finalDevs := make(map[string]*bluetooth.Device)

	for ln, l := range lights {
		uuid, err := bluetooth.ParseUUID(l.UUID)
		if err != nil {
			for _, d := range finalDevs {
				d.Disconnect()
			}
			return nil, fmt.Errorf("failed to parse uuid address[%v]: %v", l.UUID, err)
		}

		address := bluetooth.Address{
			UUID: uuid,
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

func on(dChar *bluetooth.DeviceCharacteristic) error {
	_, err := dChar.WriteWithoutResponse([]byte{0xCC, 0x23, 0x33})
	return err
}

func off(dChar *bluetooth.DeviceCharacteristic) error {
	_, err := dChar.WriteWithoutResponse([]byte{0xCC, 0x24, 0x33})
	return err
}

func setColor(dChar *bluetooth.DeviceCharacteristic, red, green, blue byte) error {
	_, err := dChar.WriteWithoutResponse([]byte{0x56, red, green, blue, 0x00, 0xF0, 0xAA})
	return err
}

func (ml *merryLighting) listenForColor() {
	for {
		select {
		case color := <-ml.colorChan:
			setColor(color.char, color.Red, color.Green, color.Blue)
		}
	}
}
